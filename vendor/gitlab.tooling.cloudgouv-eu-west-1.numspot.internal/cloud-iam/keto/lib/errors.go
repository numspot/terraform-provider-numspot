package lib

import "fmt"

// BadNamespaceError is returned when [GetNamespace] couldn't find one.
type BadNamespaceError struct {
	service     string
	resource    *string
	subresource *string
}

// Error implementation.
func (err *BadNamespaceError) Error() string {
	// Check service.
	service, ok := nameToService[err.service]
	if !ok {
		return fmt.Sprintf("service '%s' doesn't exist", err.service)
	}

	// Check resource.
	if len(service.resources) == 0 {
		return fmt.Sprintf("service '%s' doesn't have any resource", err.service)
	}

	res := make([]string, 0, len(service.resources))
	for name := range service.resources {
		res = append(res, name)
	}

	if err.resource == nil {
		return fmt.Sprintf("service '%s' exists but needs a resource. Possible choices are %v.", err.service, res)
	}

	resource, ok := service.resources[*err.resource]
	if !ok {
		return fmt.Sprintf("service '%s' doesn't have resource '%s'. Possible choices are %v", err.service, *err.resource, res)
	}

	// Check subresource.
	if len(resource.subResources) == 0 {
		return fmt.Sprintf("resource '%s' doesn't have any subresource", *err.resource)
	}

	if err.subresource == nil {
		return fmt.Sprintf("resource '%s' exists but needs a subresource. Possible choices are %v.", *err.resource, resource.SubResources())
	}

	return fmt.Sprintf("resource '%s' doesn't have subresource '%s'. Possible choices are %vV", *err.resource, *err.subresource, resource.SubResources())
}
