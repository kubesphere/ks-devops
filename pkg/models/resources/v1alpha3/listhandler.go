package v1alpha3

// ListHandler is the interface to create comparator, filter and transform
type ListHandler interface {
	Comparator() CompareFunc
	Filter() FilterFunc
	Transformer() TransformFunc
}

// defaultListHandler implements default comparator, filter and transformer.
type defaultListHandler struct {
}

func (d defaultListHandler) Comparator() CompareFunc {
	return DefaultCompare()
}

func (d defaultListHandler) Filter() FilterFunc {
	return DefaultFilter()
}

func (d defaultListHandler) Transformer() TransformFunc {
	return NoTransformFunc()
}
