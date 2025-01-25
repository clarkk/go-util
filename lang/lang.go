package lang

import (
	"log"
	"errors"
	"strconv"
	"strings"
	"github.com/clarkk/go-util/cache"
)

var (
	fetch			Adapter
	expires			int
	langs			[]string
	cache_string	*cache.Cache[string]
)

type (
	Adapter interface {
		Fetch(lang, table, key string) (string, error)
	}
	
	Lang struct {
		accept_lang	[]string
		lang		string
	}
	
	Rep 	map[string]any
)

//	Initiate with caching and fetch adapter
func Init(fetcher Adapter, cache_expires int, languages []string){
	fetch	= fetcher
	expires	= cache_expires
	langs	= make([]string, len(languages))
	for i, lang := range languages {
		langs[i] = strings.ToLower(lang)
	}
	cache_string = cache.New[string](60)
}

//	Create new instance and set language with accepted language
func New(lang string, accept_lang []string) Lang {
	l := Lang{
		accept_lang: accept_lang,
	}
	l.Set(lang)
	return l
}

//	Set language
func (l *Lang) Set(lang string){
	if lang = strings.ToLower(lang); lang != "" {
		for _, v := range langs {
			if lang == v {
				l.lang = v
				return
			}
		}
	}
	for _, a := range l.accept_lang {
		for _, v := range langs {
			if a == v {
				l.lang = v
				return
			}
		}
	}
	l.lang = langs[0]
}

//	Get string translation
func (l *Lang) String(key string, replace map[string]any) string {
	s := l.fetch("lang", key)
	if replace != nil {
		return string_replace(s, replace)
	}
	return s
}

//	Get error translation
func (l *Lang) Error(key string, replace map[string]any) error {
	s := l.fetch("lang_error", key)
	if replace != nil {
		return errors.New(string_replace(s, replace))
	}
	return errors.New(s)
}

func (l *Lang) fetch(table, key string) string {
	cache_key := l.lang+"-"+table+"-"+key
	s, ok := cache_string.Get(cache_key)
	if !ok {
		var err error
		s, err = fetch.Fetch(l.lang, table, key)
		if err != nil {
			//	Log fatal errors
			log.Printf("Lang: %v", err)
		} else {
			cache_string.Set(cache_key, s, expires)
		}
	}
	return s
}

func string_replace(s string, replace Rep) string {
	for k, v := range replace {
		switch t := v.(type) {
		case int:
			s = strings.Replace(s, "%"+k+"%", strconv.Itoa(t), -1)
		case int64:
			s = strings.Replace(s, "%"+k+"%", strconv.FormatInt(t, 10), -1)
		case float32:
			s = strings.Replace(s, "%"+k+"%", strconv.FormatFloat(float64(t), 'f', -1, 64), -1)
		case float64:
			s = strings.Replace(s, "%"+k+"%", strconv.FormatFloat(t, 'f', -1, 64), -1)
		case string:
			s = strings.Replace(s, "%"+k+"%", t, -1)
		}
	}
	return s
}