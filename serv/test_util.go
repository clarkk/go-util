package serv

func Test_set_slugs(r *http.Request, slugs ...string) *http.Request {
	ctx := context.WithValue(r.Context(), ctx_slug, slugs)
	return r.WithContext(ctx)
}