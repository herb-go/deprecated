package hiredmember

import (
	"github.com/herb-go/deprecated/member"
	"github.com/herb-go/deprecated/member-drivers/overseers/memberdirectivefactoryoverseer"
)

type Directive struct {
	ID     string
	Config func(v interface{}) error `config:", lazyload"`
}

func (d *Directive) ApplyTo(s *member.Service) error {
	f := memberdirectivefactoryoverseer.GetMemberDirectiveFactoryByID(d.ID)
	directive, err := f(d.Config)
	if err != nil {
		return err
	}
	return directive.Execute(s)
}

type Config struct {
	Directives []*Directive
}

func (c *Config) ApplyTo(s *member.Service) error {
	for k := range c.Directives {
		err := c.Directives[k].ApplyTo(s)
		if err != nil {
			return err
		}
	}
	return nil
}
