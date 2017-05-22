package dcmd

type Trigger struct {
	Names       []string
	Middlewares []MiddleWareFunc

	HideFromHelp     bool
	DisableInDM      bool
	DisableOutsideDM bool
}

func NewTrigger(name string, aliases ...string) *Trigger {
	names := []string{name}
	if len(aliases) > 0 {
		names = append(names, aliases...)
	}

	return &Trigger{
		Names: names,
	}
}

func (t *Trigger) SetHideFromHelp(hide bool) *Trigger {
	t.HideFromHelp = hide
	return t
}

func (t *Trigger) SetDisableInDM(disable bool) *Trigger {
	t.DisableInDM = disable
	return t
}

func (t *Trigger) SetDisableOutsideDM(disable bool) *Trigger {
	t.DisableOutsideDM = disable
	return t
}

func (t *Trigger) SetMiddlewares(mw ...MiddleWareFunc) *Trigger {
	t.Middlewares = mw
	return t
}
