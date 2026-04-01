package env

import (
	"context"
	"net/http"
	"github.com/clarkk/go-util/lang"
	"github.com/clarkk/go-util/serv/req"
	"github.com/clarkk/go-util/sess"
)

const ctx_env ctx_key = ""

type (
	Env_data interface {
		Session() *sess.Session
		Lang() string
		Data() Properties
		Data_session() Properties
		Update_lang(string)
		Update(map[string]any) error
	}
	
	Environment struct {
		Env_data
		r		*http.Request
		lang	lang.Lang
	}
	
	Properties map[string]any
	
	ctx_key		string
)

func New(d Env_data) *Environment {
	e := &Environment{
		Env_data:	d,
	}
	e.Set_lang(d.Lang())
	return e
}

func New_request(r *http.Request, d Env_data) *Environment {
	e := &Environment{
		Env_data:	d,
		r:			r,
	}
	e.Set_lang(d.Lang())
	
	ctx := context.WithValue(r.Context(), ctx_env, e)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	return e
}

func Request(r *http.Request) *Environment {
	e, _ := r.Context().Value(ctx_env).(*Environment)
	return e
}

func (e *Environment) Set_lang(l string){
	var accept_lang []string
	if l == "" && e.r != nil {
		accept_lang = req.Accept_lang(e.r)
	}
	e.lang = lang.New(l, accept_lang)
	e.Env_data.Update_lang(e.lang.Get())
}

func (e *Environment) Lang() string {
	return e.lang.Get()
}

func (e *Environment) Lang_string(key string, replace map[string]any) string {
	return e.lang.String(key, replace)
}

func (e *Environment) Lang_error(key string, replace map[string]any) error {
	return e.lang.Error(key, replace)
}

func (e *Environment) Lang_printer() *lang.Printer {
	return e.lang.Printer()
}