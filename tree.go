package web

// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
// https://github.com/julienschmidt/httprouter/blob/master/tree.go

import (
	"strings"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// Search for a wildcard segment and check the name for invalid characters.
// Returns -1 as index, if no wildcard was found.
func findWildcard(path string) (wilcard string, i int, valid bool) {
	// Find start
	for start := 0; start < len(path); start++ {
		c := path[start]
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end := start + 1; end < len(path); end++ {
			c = path[end]
			switch c {
			case '/':
				return path[start:end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

func countParams(path string) uint16 {
	var n uint
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case ':', '*':
			n++
		}
	}
	return uint16(n)
}

type nodeType uint8

const (
	static nodeType = iota // default
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node
	next      Next
}

type frozenNode struct {
	path           string
	nType          nodeType
	staticIndices  string
	staticChildren []*frozenNode
	paramChild     *frozenNode
	catchAllChild  *frozenNode
	nextChild      *frozenNode
	route          *frozenRoute
}

type frozenRoute struct {
	paramNames []string
	next       Next
}

type skippedNode struct {
	node      *node
	path      string
	paramsLen int
}

type frozenSkippedNode struct {
	node      *frozenNode
	path      string
	paramsLen int
}

type routeParams struct {
	names       []string
	extraValues *[]string
	count       uint16
	value0      string
	value1      string
	value2      string
}

func (p *routeParams) addValue(app *Application, value string) {
	switch p.count {
	case 0:
		p.value0 = value
	case 1:
		p.value1 = value
	case 2:
		p.value2 = value
	default:
		if app != nil {
			if p.extraValues == nil {
				p.extraValues = app.getParamValues()
			}
			i := int(p.count) - 3
			*p.extraValues = (*p.extraValues)[:i+1]
			(*p.extraValues)[i] = value
		}
	}
	p.count++
}

func (p *routeParams) truncate(paramsLen int) {
	p.count = uint16(paramsLen)
	if p.extraValues != nil {
		extraLen := paramsLen - 3
		if extraLen < 0 {
			extraLen = 0
		}
		*p.extraValues = (*p.extraValues)[:extraLen]
	}
}

func (p *routeParams) valueAt(i int) string {
	switch i {
	case 0:
		return p.value0
	case 1:
		return p.value1
	case 2:
		return p.value2
	default:
		return (*p.extraValues)[i-3]
	}
}

type routeMatch struct {
	callback Next
	params   routeParams
	tsr      bool
}

func (n *node) wildcardChild() *node {
	if !n.wildChild {
		return nil
	}
	return n.children[len(n.children)-1]
}

func (n *node) addStaticChild(idxc byte) *node {
	child := &node{}
	n.indices += string([]byte{idxc})

	if n.wildChild {
		wildcardIndex := len(n.children) - 1
		n.children = append(n.children, nil)
		copy(n.children[wildcardIndex+1:], n.children[wildcardIndex:])
		n.children[wildcardIndex] = child
	} else {
		n.children = append(n.children, child)
	}

	pos := n.incrementChildPrio(len(n.indices) - 1)
	return n.children[pos]
}

func isCatchAllPlaceholder(n *node) bool {
	return n != nil &&
		n.nType == catchAll &&
		n.path == "" &&
		n.wildChild &&
		len(n.children) == 1 &&
		n.children[0] != nil &&
		n.children[0].nType == catchAll
}

func (n *node) hasStaticParamSibling() bool {
	if n == nil {
		return false
	}
	if n.wildChild && len(n.indices) > 0 {
		if child := n.wildcardChild(); child != nil && child.nType == param {
			return true
		}
	}
	for i := 0; i < len(n.children); i++ {
		if n.children[i].hasStaticParamSibling() {
			return true
		}
	}
	return false
}

func (n *node) freeze() *frozenNode {
	return n.freezeWithParams(nil)
}

func cloneStrings(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

func (n *node) freezeWithParams(paramNames []string) *frozenNode {
	if n == nil {
		return nil
	}

	currentParamNames := paramNames
	switch n.nType {
	case param:
		currentParamNames = append(cloneStrings(paramNames), n.path[1:])
	case catchAll:
		if len(n.path) > 2 {
			currentParamNames = append(cloneStrings(paramNames), n.path[2:])
		}
	}

	f := &frozenNode{
		path:  n.path,
		nType: n.nType,
	}
	if n.next != nil {
		f.route = &frozenRoute{
			paramNames: cloneStrings(currentParamNames),
			next:       n.next,
		}
	}

	if n.nType == param {
		if len(n.children) > 0 {
			f.nextChild = n.children[0].freezeWithParams(currentParamNames)
		}
		return f
	}

	for i := 0; i < len(n.indices); i++ {
		child := n.children[i]
		if isCatchAllPlaceholder(child) && child.children[0] != nil {
			f.catchAllChild = child.children[0].freezeWithParams(currentParamNames)
			continue
		}
		f.staticIndices += n.indices[i : i+1]
		f.staticChildren = append(f.staticChildren, child.freezeWithParams(currentParamNames))
	}

	if n.wildChild {
		child := n.wildcardChild()
		switch child.nType {
		case param:
			f.paramChild = child.freezeWithParams(currentParamNames)
		case catchAll:
			f.catchAllChild = child.freezeWithParams(currentParamNames)
		}
	}

	return f
}

func (n *frozenNode) staticChild(idxc byte) *frozenNode {
	switch len(n.staticIndices) {
	case 0:
		return nil
	case 1:
		if n.staticIndices[0] == idxc {
			return n.staticChildren[0]
		}
		return nil
	case 2:
		if n.staticIndices[0] == idxc {
			return n.staticChildren[0]
		}
		if n.staticIndices[1] == idxc {
			return n.staticChildren[1]
		}
		return nil
	case 3:
		if n.staticIndices[0] == idxc {
			return n.staticChildren[0]
		}
		if n.staticIndices[1] == idxc {
			return n.staticChildren[1]
		}
		if n.staticIndices[2] == idxc {
			return n.staticChildren[2]
		}
		return nil
	}
	for i := 0; i < len(n.staticIndices); i++ {
		if n.staticIndices[i] == idxc {
			return n.staticChildren[i]
		}
	}
	return nil
}

func (n *frozenNode) lookup(path string, app *Application, match *routeMatch) {
	var skippedNode0 *frozenNode
	var skippedPath0 string
	var skippedParamsLen0 int
	var skippedMore []frozenSkippedNode
	skippedLen := 0
	forceWildcard := false

walk:
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				search := path
				path = path[len(prefix):]

				idxc := path[0]
				if !forceWildcard {
					var child *frozenNode
					switch len(n.staticIndices) {
					case 1:
						if n.staticIndices[0] == idxc {
							child = n.staticChildren[0]
						}
					case 2:
						if n.staticIndices[0] == idxc {
							child = n.staticChildren[0]
						} else if n.staticIndices[1] == idxc {
							child = n.staticChildren[1]
						}
					case 3:
						if n.staticIndices[0] == idxc {
							child = n.staticChildren[0]
						} else if n.staticIndices[1] == idxc {
							child = n.staticChildren[1]
						} else if n.staticIndices[2] == idxc {
							child = n.staticChildren[2]
						}
					default:
						for i := 0; i < len(n.staticIndices); i++ {
							if n.staticIndices[i] == idxc {
								child = n.staticChildren[i]
								break
							}
						}
					}
					if child != nil {
						if n.paramChild != nil {
							paramsLen := int(match.params.count)
							if skippedLen == 0 {
								skippedNode0 = n
								skippedPath0 = search
								skippedParamsLen0 = paramsLen
							} else {
								skippedMore = append(skippedMore, frozenSkippedNode{
									node:      n,
									path:      search,
									paramsLen: paramsLen,
								})
							}
							skippedLen++
						}
						n = child
						continue walk
					}
				}
				forceWildcard = false

				if n.paramChild != nil {
					n = n.paramChild

					end := strings.IndexByte(path, '/')
					if end < 0 {
						end = len(path)
					}

					match.params.addValue(app, path[:end])

					if end < len(path) {
						if n.nextChild != nil {
							path = path[end:]
							n = n.nextChild
							continue walk
						}
						if skippedLen > 0 {
							goto backtrack
						}
						match.tsr = (len(path) == end+1)
						return
					}

					if n.route != nil {
						match.callback = n.route.next
						match.params.names = n.route.paramNames
						return
					}
					if n.nextChild != nil {
						match.tsr = (n.nextChild.path == "/" && n.nextChild.route != nil) ||
							(n.nextChild.path == "" && n.nextChild.staticChild('/') != nil)
					}
					if skippedLen > 0 {
						goto backtrack
					}
					return
				}

				if n.catchAllChild != nil {
					n = n.catchAllChild
					match.params.addValue(app, path)
					match.params.names = n.route.paramNames
					match.callback = n.route.next
					return
				}

				if skippedLen > 0 {
					goto backtrack
				}
				match.tsr = (path == "/" && n.route != nil)
				return
			}
		} else if path == prefix {
			if n.route != nil {
				match.callback = n.route.next
				match.params.names = n.route.paramNames
				return
			}

			if path == "/" && (n.paramChild != nil || n.catchAllChild != nil) && n.nType != root {
				match.tsr = true
				return
			}

			if path == "/" && n.nType == static {
				match.tsr = true
				return
			}

			if child := n.staticChild('/'); child != nil {
				match.tsr = (len(child.path) == 1 && child.route != nil) || child.catchAllChild != nil
				return
			}
			if n.catchAllChild != nil {
				match.tsr = true
				return
			}
			if skippedLen > 0 {
				goto backtrack
			}
			return
		}

		if skippedLen > 0 {
			goto backtrack
		}
		match.tsr = (path == "/") ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.route != nil)
		return
	}

backtrack:
	skippedLen--
	if skippedLen == 0 {
		n = skippedNode0
		path = skippedPath0
		match.params.truncate(skippedParamsLen0)
	} else {
		last := len(skippedMore) - 1
		skippedNode := skippedMore[last]
		skippedMore = skippedMore[:last]
		n = skippedNode.node
		path = skippedNode.path
		match.params.truncate(skippedNode.paramsLen)
	}
	forceWildcard = true
	match.tsr = false
	goto walk
}

func (n *frozenNode) getValue(path string, app *Application) (callback Next, ps *Params, tsr bool) {
	var match routeMatch
	n.lookup(path, app, &match)
	callback = match.callback
	tsr = match.tsr
	if callback == nil || app == nil || match.params.count == 0 {
		return
	}

	ps = app.getParams()
	*ps = (*ps)[:match.params.count]
	for i := 0; i < int(match.params.count); i++ {
		(*ps)[i] = Param{
			Key:   match.params.names[i],
			Value: match.params.valueAt(i),
		}
	}
	app.putParamValues(match.params.extraValues)
	return
}

func (n *node) getValueFast(path string, app *Application) (callback Next, ps *Params, tsr bool) {
walk:
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if strings.HasPrefix(path, prefix) {
				path = path[len(prefix):]

				if !n.wildChild {
					idxc := path[0]
					for i := 0; i < len(n.indices); i++ {
						if n.indices[i] == idxc {
							n = n.children[i]
							continue walk
						}
					}

					tsr = (path == "/" && n.next != nil)
					return
				}

				n = n.children[0]
				switch n.nType {
				case param:
					end := strings.IndexByte(path, '/')
					if end < 0 {
						end = len(path)
					}

					if app != nil {
						if ps == nil {
							ps = app.getParams()
						}
						i := len(*ps)
						*ps = (*ps)[:i+1]
						(*ps)[i] = Param{
							Key:   n.path[1:],
							Value: path[:end],
						}
					}

					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						tsr = (len(path) == end+1)
						return
					}

					if callback = n.next; callback != nil {
						return
					} else if len(n.children) == 1 {
						n = n.children[0]
						tsr = (n.path == "/" && n.next != nil) || (n.path == "" && n.indices == "/")
					}

					return

				case catchAll:
					if app != nil {
						if ps == nil {
							ps = app.getParams()
						}
						i := len(*ps)
						*ps = (*ps)[:i+1]
						(*ps)[i] = Param{
							Key:   n.path[2:],
							Value: path,
						}
					}

					callback = n.next
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == prefix {
			if callback = n.next; callback != nil {
				return
			}

			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}

			if path == "/" && n.nType == static {
				tsr = true
				return
			}

			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == '/' {
					n = n.children[i]
					tsr = (len(n.path) == 1 && n.next != nil) ||
						(n.nType == catchAll && n.children[0].next != nil)
					return
				}
			}
			return
		}

		tsr = (path == "/") ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.next != nil)
		return
	}
}

// Increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // Unchanged prefix, might be empty
			n.indices[pos:pos+1] + // The index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // Rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given callback to the path.
// Not concurrency-safe!
func (n *node) addRoute(path string, callback Next) {
	fullPath := path
	n.priority++

	// Empty tree
	if n.path == "" && n.indices == "" {
		n.insertChild(path, fullPath, callback)
		n.nType = root
		return
	}

walk:
	for {
		// Find the longest common prefix.
		// This also implies that the common prefix contains no ':' or '*'
		// since the existing key can't contain those chars.
		i := longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path:      n.path[i:],
				wildChild: n.wildChild,
				nType:     static,
				indices:   n.indices,
				children:  n.children,
				next:      n.next,
				priority:  n.priority - 1,
			}

			n.children = []*node{&child}
			// []byte for proper unicode char conversion, see #65
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.next = nil
			n.wildChild = false
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]
			idxc := path[0]

			if n.wildChild {
				if idxc != ':' && idxc != '*' {
					// Match static children before the wildcard child so static
					// siblings stay reachable regardless of registration order.
					for i := 0; i < len(n.indices); i++ {
						if n.indices[i] == idxc {
							i = n.incrementChildPrio(i)
							n = n.children[i]
							continue walk
						}
					}

					if n.wildcardChild().nType != catchAll {
						n = n.addStaticChild(idxc)
						n.insertChild(path, fullPath, callback)
						return
					}
				}

				n = n.wildcardChild()
				n.priority++

				// Check if the wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					// Adding a child to a catchAll is not possible
					n.nType != catchAll &&
					// Check for longer wildcard, e.g. :name and :names
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				}

				// Wildcard conflict
				pathSeg := path
				if n.nType != catchAll {
					pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
				}
				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
				panic("'" + pathSeg +
					"' in new path '" + fullPath +
					"' conflicts with existing wildcard '" + n.path +
					"' in existing prefix '" + prefix +
					"'")
			}

			// '/' after param
			if n.nType == param && idxc == '/' && len(n.children) == 1 {
				n = n.children[0]
				n.priority++
				continue walk
			}

			// Check if a child with the next path byte exists
			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == idxc {
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Otherwise insert it
			if idxc != ':' && idxc != '*' {
				n = n.addStaticChild(idxc)
			}
			n.insertChild(path, fullPath, callback)
			return
		}

		// Otherwise add callback to current node
		if n.next != nil {
			panic("a callback is already registered for path '" + fullPath + "'")
		}
		n.next = callback
		return
	}
}

func (n *node) insertChild(path, fullPath string, callback Next) {
	for {
		// Find prefix until first wildcard
		wildcard, i, valid := findWildcard(path)
		if i < 0 { // No wilcard found
			break
		}

		// The wildcard name must not contain ':' and '*'
		if !valid {
			panic("only one wildcard per path segment is allowed, has: '" +
				wildcard + "' in path '" + fullPath + "'")
		}

		// Check if the wildcard has a name
		if len(wildcard) < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		// Check if this node has existing children which would be
		// unreachable if we insert the wildcard here.
		// Param children may coexist with existing static children.
		if len(n.children) > 0 && (wildcard[0] != ':' || n.wildChild) {
			panic("wildcard segment '" + wildcard +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		// param
		if wildcard[0] == ':' {
			if i > 0 {
				// Insert prefix before the current wildcard
				n.path = path[:i]
				path = path[i:]
			}

			n.wildChild = true
			child := &node{
				nType: param,
				path:  wildcard,
			}
			if len(n.children) == 0 {
				n.children = []*node{child}
			} else {
				n.children = append(n.children, child)
			}
			n = child
			n.priority++

			// If the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]
				child := &node{
					priority: 1,
				}
				n.children = []*node{child}
				n = child
				continue
			}

			// Otherwise we're done. Insert the callback in the new leaf
			n.next = callback
			return
		}

		// catchAll
		if i+len(wildcard) != len(path) {
			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
		}

		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
			panic("catch-all conflicts with existing callback for the path segment root in path '" + fullPath + "'")
		}

		// Currently fixed width 1 for '/'
		i--
		if path[i] != '/' {
			panic("no / before catch-all in path '" + fullPath + "'")
		}

		n.path = path[:i]

		// First node: catchAll node with empty path
		child := &node{
			wildChild: true,
			nType:     catchAll,
		}
		n.children = []*node{child}
		n.indices = string('/')
		n = child
		n.priority++

		// Second node: node holding the variable
		child = &node{
			path:     path[i:],
			nType:    catchAll,
			next:     callback,
			priority: 1,
		}
		n.children = []*node{child}

		return
	}

	// If no wildcard was found, simply insert the path and callback
	n.path = path
	n.next = callback
}

// Returns the callback registered with the given path (key). The values of
// wildcards are saved to a map.
// If no callback can be found, a TSR (trailing slash redirect) recommendation is
// made if a callback exists with an extra (without the) trailing slash for the
// given path.
func (n *node) getValue(path string, app *Application) (callback Next, ps *Params, tsr bool) {
	var skippedNode0 *node
	var skippedPath0 string
	var skippedParamsLen0 int
	var skippedMore []skippedNode
	skippedLen := 0
	forceWildcard := false

walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				search := path
				path = path[len(prefix):]

				idxc := path[0]
				if !forceWildcard {
					for i := 0; i < len(n.indices); i++ {
						if n.indices[i] == idxc {
							if n.wildChild {
								paramsLen := 0
								if ps != nil {
									paramsLen = len(*ps)
								}
								if skippedLen == 0 {
									skippedNode0 = n
									skippedPath0 = search
									skippedParamsLen0 = paramsLen
								} else {
									skippedMore = append(skippedMore, skippedNode{
										node:      n,
										path:      search,
										paramsLen: paramsLen,
									})
								}
								skippedLen++
							}
							n = n.children[i]
							continue walk
						}
					}
				}
				forceWildcard = false

				// If this node does not have a wildcard (param or catchAll)
				// child, we can just look up the next child node and continue
				// to walk down the tree
				if !n.wildChild {
					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					if skippedLen > 0 {
						goto backtrack
					}
					tsr = (path == "/" && n.next != nil)
					return
				}

				// Callback wildcard child
				n = n.wildcardChild()
				switch n.nType {
				case param:
					// Find param end (either '/' or path end)
					end := strings.IndexByte(path, '/')
					if end < 0 {
						end = len(path)
					}

					// Save param value
					if app != nil {
						if ps == nil {
							ps = app.getParams()
						}
						// Expand slice within preallocated capacity
						i := len(*ps)
						*ps = (*ps)[:i+1]
						(*ps)[i] = Param{
							Key:   n.path[1:],
							Value: path[:end],
						}
					}

					// We need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						if skippedLen > 0 {
							goto backtrack
						}
						tsr = (len(path) == end+1)
						return
					}

					if callback = n.next; callback != nil {
						return
					} else if len(n.children) == 1 {
						// No callback found. Check if a callback for this path + a
						// trailing slash exists for TSR recommendation
						n = n.children[0]
						tsr = (n.path == "/" && n.next != nil) || (n.path == "" && n.indices == "/")
					}

					if skippedLen > 0 {
						goto backtrack
					}
					return

				case catchAll:
					// Save param value
					if app != nil {
						if ps == nil {
							ps = app.getParams()
						}
						// Expand slice within preallocated capacity
						i := len(*ps)
						*ps = (*ps)[:i+1]
						(*ps)[i] = Param{
							Key:   n.path[2:],
							Value: path,
						}
					}

					callback = n.next
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == prefix {
			// We should have reached the node containing the callback.
			// Check if this node has a callback registered.
			if callback = n.next; callback != nil {
				return
			}

			// If there is no callback for this route, but this route has a
			// wildcard child, there must be a callback for this path with an
			// additional trailing slash
			if path == "/" && n.wildChild && n.nType != root {
				tsr = true
				return
			}

			if path == "/" && n.nType == static {
				tsr = true
				return
			}

			// No callback found. Check if a callback for this path + a
			// trailing slash exists for trailing slash recommendation
			for i := 0; i < len(n.indices); i++ {
				if n.indices[i] == '/' {
					n = n.children[i]
					tsr = (len(n.path) == 1 && n.next != nil) ||
						(n.nType == catchAll && n.children[0].next != nil)
					return
				}
			}
			if skippedLen > 0 {
				goto backtrack
			}
			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		if skippedLen > 0 {
			goto backtrack
		}
		tsr = (path == "/") ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.next != nil)
		return
	}

backtrack:
	skippedLen--
	if skippedLen == 0 {
		n = skippedNode0
		path = skippedPath0
		if ps != nil {
			*ps = (*ps)[:skippedParamsLen0]
		}
	} else {
		last := len(skippedMore) - 1
		skippedNode := skippedMore[last]
		skippedMore = skippedMore[:last]
		n = skippedNode.node
		path = skippedNode.path
		if ps != nil {
			*ps = (*ps)[:skippedNode.paramsLen]
		}
	}
	forceWildcard = true
	tsr = false
	goto walk
}
