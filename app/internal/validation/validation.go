package validation

type Localizer interface {
	Format(string, map[string]any) string
	T(string) string
}

type Field struct {
	Name  string
	Error FormattableError
}

type List []Field

func With(name string, error FormattableError) List {
	return List{Field{Name: name, Error: error}}
}

func (l List) None() bool {
	return len(l) == 0
}

func (l List) Any() bool {
	return len(l) > 0
}

func (l List) With(name string, error FormattableError) List {
	if l.Has(name) {
		return l
	}

	return append(l, Field{Name: name, Error: error})
}

func (l *List) Add(name string, error FormattableError) {
	if l.Has(name) {
		return
	}

	*l = append(*l, Field{Name: name, Error: error})
}

func (l List) Has(name string) bool {
	for _, field := range l {
		if field.Name == name {
			return true
		}
	}

	return false
}

func (l List) Format(localizer Localizer, name string) string {
	for _, field := range l {
		if field.Name == name {
			return field.Error.Format(localizer)
		}
	}

	return ""
}
