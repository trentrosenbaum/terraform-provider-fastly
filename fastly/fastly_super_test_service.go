package fastly

// SuperTestServiceMetadata provides a container to pass metadata about services into SuperTests and SuperTestComponents
type SuperTestService struct {
	Service SuperTestServiceMetadata
}

// SuperTestServiceMetadata provides that actual metadata about services.
// It is a sub-struct so that when mixed into SuperTests and SuperTestComponents they are referenced under "Service.*".
type SuperTestServiceMetadata struct {
	Type         string
	ResourceId   string
	TemplateName string
}
