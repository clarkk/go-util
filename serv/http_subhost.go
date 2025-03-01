package serv

import (
	"log"
	"strings"
	"regexp"
	"net/http"
)

const (
	ALL Method 		= "*"
	GET Method		= "GET"
	POST Method		= "POST"
	DELETE Method	= "DELETE"
	
	pattern_slug	= ":slug"
	pattern_file 	= ":file"
	
	re_slug_pattern = `([^/]+)`
	re_file_pattern = `([a-z\d\-_]+\.[a-z]{1,4})`
)

var (
	re_slug 		= regexp.MustCompile(`^[\p{L}\d.\-_]*$`)
)

type (
	Method 			string
	
	subhost struct {
		map_routes 	map_routes
		map_exact	map_exact
		routes 		routes
	}
	
	map_routes 		map[string]route_handlers
	map_exact		map[string]bool
	routes 			[]*route
	
	route struct {
		route_pattern
		methods 	route_handlers
	}
	
	route_pattern struct {
		pattern 	string
		exact		bool
		slugs 		[]string
		depth 		int
		regex 		*regexp.Regexp
	}
	
	route_handlers 	map[string]*route_handler
	
	route_handler struct {
		timeout 	int
		blind		bool
		handler		http.HandlerFunc
	}
)

//	Apply route pattern exact
func (s *subhost) Route_exact(method Method, pattern string, timeout int, handler http.HandlerFunc) *subhost {
	return s.route(method, pattern, timeout, handler, true, false)
}

//	Apply route pattern
func (s *subhost) Route(method Method, pattern string, timeout int, handler http.HandlerFunc) *subhost {
	return s.route(method, pattern, timeout, handler, false, false)
}

//	Apply blind route pattern (HTTP 404)
func (s *subhost) Route_blind(method Method, pattern string) *subhost {
	var handler http.HandlerFunc
	return s.route(method, pattern, 0, handler, false, true)
}

func (s *subhost) route(method Method, pattern string, timeout int, handler http.HandlerFunc, exact, blind bool) *subhost {
	validate_pattern(pattern)
	
	key_method 	:= string(method)
	timeout 	= timeout_min(timeout)
	
	if existing_route, ok := s.map_routes[pattern]; ok {
		s.validate_existing_route(method, pattern, exact, existing_route)
		
		existing_route[key_method] = &route_handler{
			timeout:	timeout,
			blind:		blind,
			handler:	handler,
		}
	} else {
		methods := route_handlers{
			key_method: &route_handler{
				timeout:	timeout,
				blind:		blind,
				handler:	handler,
			},
		}
		
		s.map_routes[pattern]	= methods
		s.map_exact[pattern]	= exact
		s.routes = append(s.routes, &route{
			route_pattern:	parse_route_pattern(pattern, exact),
			methods:		methods,
		})
	}
	
	return s
}

func (s *subhost) validate_existing_route(method Method, pattern string, exact bool, existing_route route_handlers){
	if _, ok := existing_route[string(method)]; ok {
		log.Fatalf("Route is duplicate: %s %s", method, pattern)
	}
	
	if method == ALL {
		log.Fatalf("Route is redundant: %s %s", method, pattern)
	} else if _, ok := existing_route[string(ALL)]; ok {
		log.Fatalf("Route is redundant: %s %s", method, pattern)
	}
	
	if s.map_exact[pattern] != exact {
		log.Fatal("Routes with exact/prefix can not be mixed")
	}
}

func parse_route_pattern(pattern string, exact bool) route_pattern {
	pattern = strip_trailing_slash(pattern)
	
	if pattern == "/" {
		return route_pattern{
			pattern:	pattern,
			exact:		exact,
		}
	}
	
	var (
		re 			string
		has_regex 	bool
		regex 		*regexp.Regexp
	)
	
	slugs := strings.Split(strings.TrimLeft(pattern, "/"), "/")
	depth := len(slugs)
	
	for i, slug := range slugs {
		if slug == "" {
			log.Fatalf("Route slug can not be empty: %s", pattern)
		}
		
		if slug[0] == ':' {
			has_regex = true
			switch slug {
			case pattern_slug:
				re += "/"+re_slug_pattern
			case pattern_file:
				if !exact || depth-1 != i {
					log.Fatalf("Route file can only be the last level in combination with exact: %s", pattern)
				}
				re += "/"+re_file_pattern
			default:
				log.Fatalf("Invalid regex parameter: %s", slug)
			}
		} else {
			if !re_slug.MatchString(slug) {
				log.Fatalf("Invalid chars in slug: %s (%s)", slug, pattern)
			}
			
			re += "/"+slug
		}
	}
	
	if has_regex {
		if exact {
			regex = regexp.MustCompile("^"+re+"$")
		} else {
			regex = regexp.MustCompile("^"+re)
		}
	}
	
	return route_pattern{
		pattern:	pattern,
		exact:		exact,
		slugs:		slugs,
		depth:		depth,
		regex:		regex,
	}
}

func validate_pattern(pattern string){
	if pattern == "" {
		log.Fatal("Route cannot be empty")
	}
	if pattern[0] != '/' {
		log.Fatalf("Route must start with '/': %s -> /%s", pattern, pattern)
	}
}

func timeout_min(timeout int) int {
	if timeout < 0 {
		return 0
	}
	return timeout
}