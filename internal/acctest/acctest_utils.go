package acctest

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

const (
	// PairPrefix handles pair prefix.
	PairPrefix    = "[[FROM PAIR]]"
	sentinelIndex = "*"
)

var (
	// errEmptyIDField is returned when a Terraform resource ID field is empty.
	errEmptyIDField = errors.New("id field should not be empty")
	// errIDUnchanged is returned when a resource ID does not match the expected
	// value.
	errIDUnchanged = errors.New("id should be unchanged")
	// errIDChanged is returned when a resource ID was expected to change but did
	// not.
	errIDChanged = errors.New("id should have changed")
	// errNoNonEmptyValues is returned when a set of expected values is empty.
	errNoNonEmptyValues = errors.New("has no non-empty values")
	// errComparingSelf is returned when a test compares a resource with itself.
	errComparingSelf = errors.New("comparing self")
	// errAttributeNotSet is returned when an expected attribute is missing.
	errAttributeNotSet = errors.New("attribute not set")
	// errAttributeMismatch is returned when two paired attributes do not match.
	errAttributeMismatch = errors.New("attribute mismatch")
	// errResourceNotFound is returned when a Terraform state resource cannot be
	// found.
	errResourceNotFound = errors.New("not found")
	// errNoPrimaryInstance is returned when a Terraform state resource has no
	// primary instance.
	errNoPrimaryInstance = errors.New("no primary instance")
	// errMissingSpecialSuffix is returned when an attribute does not end with the
	// sentinel suffix.
	errMissingSpecialSuffix = errors.New("does not end with the special value")
	// errInvalidPairFieldName is returned when a paired-field name is malformed.
	errInvalidPairFieldName = errors.New("invalid field name")
	// errTypeSetElementNotFound is returned when a TypeSet element is missing.
	errTypeSetElementNotFound = errors.New("no TypeSet element")
)

// GetEnvOrReplay returns the value of envName. In replay mode the env var is
// ignored and replayValue is returned instead. Outside replay mode, the env
// var is required and the test fails if it is empty.
func GetEnvOrReplay(t *testing.T, envName, replayValue string) string {
	t.Helper()
	if vcrMode() == vcrModeReplay {
		return replayValue
	}

	v := os.Getenv(envName)
	if v == "" {
		t.Fatalf("%s environment variable must be set when VCR_MODE is not replay", envName)
	}
	return v
}

// InitResourceId handles init resource id.
func InitResourceId(t *testing.T, v string, resourceId *string) error {
	if !assert.NotEmpty(t, v) {
		return errEmptyIDField
	}
	*resourceId = v
	return nil
}

// CheckResourceIdUnchanged handles check resource id unchanged.
func CheckResourceIdUnchanged(t *testing.T, v string, resourceId *string) error {
	if !assert.NotEmpty(t, v) {
		return errEmptyIDField
	}
	if !assert.Equal(t, *resourceId, v) {
		return fmt.Errorf("%w. Expected %s but got %s", errIDUnchanged, *resourceId, v)
	}
	return nil
}

// CheckResourceIdChanged handles check resource id changed.
func CheckResourceIdChanged(t *testing.T, v string, resourceId *string) error {
	if !assert.NotEmpty(t, v) {
		return errEmptyIDField
	}
	if !assert.NotEqual(t, *resourceId, v) {
		return errIDChanged
	}
	*resourceId = v
	return nil
}

// func ListToStringList(list []string) string {
//	var strList string
//	if len(list) > 0 {
//		strList = fmt.Sprintf(`["%s"]`, strings.Join(list, `", "`))
//	} else {
//		strList = `[]`
//	}
//
//	return strList
//}

// TestCheckTypeSetElemNestedAttrsWithPair handles test check type set elem nested attrs with pair.
func TestCheckTypeSetElemNestedAttrsWithPair(name, attr string, values map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// account for cases where the user is trying to see if the value is unset/empty
		// there may be ambiguous scenarios where a field was deliberately unset vs set
		// to the empty string, this will match both, which may be a false positive.
		var matchCount int
		for _, v := range values {
			if v != "" {
				matchCount++
			}
		}
		if matchCount == 0 {
			return fmt.Errorf("%w: %#v has no non-empty values", errNoNonEmptyValues, values)
		}

		return testCheckTypeSetElemNestedAttrsInStateWithPair(s, name, attr, matchCount, values)
	}
}

// private function forked and unchanged from github.com/hashicorp/terraform-plugin-testing/helper/resource
func testCheckResourceAttrPair(
	isFirst *terraform.InstanceState,
	nameFirst string,
	keyFirst string,
	isSecond *terraform.InstanceState,
	nameSecond string,
	keySecond string,
) error {
	if nameFirst == nameSecond && keyFirst == keySecond {
		return fmt.Errorf(
			"%w: resource %s attribute %s",
			errComparingSelf,
			nameFirst,
			keyFirst,
		)
	}

	vFirst, okFirst := isFirst.Attributes[keyFirst]
	vSecond, okSecond := isSecond.Attributes[keySecond]

	okFirst, okSecond = normalizeCollectionCountKeys(keyFirst, keySecond, vFirst, vSecond, okFirst, okSecond)

	return compareAttributeValues(nameFirst, keyFirst, vFirst, okFirst, nameSecond, keySecond, vSecond, okSecond)
}

// normalizeCollectionCountKeys treats container count values of 0 and empty as
// unset so that comparisons between count keys are not sensitive to schema
// normalization differences.
func normalizeCollectionCountKeys(keyFirst, keySecond, vFirst, vSecond string, okFirst, okSecond bool) (okFirstOut, okSecondOut bool) {
	if !isMatchingCollectionCountKey(keyFirst, keySecond) {
		return okFirst, okSecond
	}

	if vFirst == "0" || vFirst == "" {
		okFirst = false
	}
	if vSecond == "0" || vSecond == "" {
		okSecond = false
	}

	return okFirst, okSecond
}

func isMatchingCollectionCountKey(keyFirst, keySecond string) bool {
	if len(keyFirst) <= 2 || len(keySecond) <= 2 {
		return false
	}

	return keyFirst[len(keyFirst)-2:] == keySecond[len(keySecond)-2:] &&
		(strings.HasSuffix(keyFirst, ".#") || strings.HasSuffix(keyFirst, ".%"))
}

func compareAttributeValues(
	nameFirst, keyFirst, vFirst string,
	okFirst bool,
	nameSecond, keySecond, vSecond string,
	okSecond bool,
) error {
	if okFirst != okSecond {
		return attributePresenceMismatch(nameFirst, keyFirst, vFirst, okFirst, nameSecond, keySecond, vSecond)
	}

	if !okFirst && !okSecond {
		return nil
	}

	if vFirst != vSecond {
		return fmt.Errorf(
			"%w: %s: Attribute '%s' expected %#v, got %#v",
			errAttributeMismatch,
			nameFirst,
			keyFirst,
			vSecond,
			vFirst)
	}

	return nil
}

func attributePresenceMismatch(
	nameFirst, keyFirst, vFirst string,
	okFirst bool,
	nameSecond, keySecond, vSecond string,
) error {
	if !okFirst {
		return fmt.Errorf(
			"%w: %s: Attribute %q not set, but %q is set in %s as %q",
			errAttributeNotSet,
			nameFirst,
			keyFirst,
			keySecond,
			nameSecond,
			vSecond,
		)
	}

	return fmt.Errorf(
		"%w: %s: Attribute %q is %q, but %q is not set in %s",
		errAttributeMismatch,
		nameFirst,
		keyFirst,
		vFirst,
		keySecond,
		nameSecond,
	)
}

// private function forked and unchanged from github.com/hashicorp/terraform-plugin-testing/helper/resource
// primaryInstanceState returns the primary instance state for the given
// resource name in the root module.
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule()
	return modulePrimaryInstanceState(ms, name)
}

// private function forked and unchanged from github.com/hashicorp/terraform-plugin-testing/helper/resource
// modulePrimaryInstanceState returns the instance state for the given resource
// name in a ModuleState
func modulePrimaryInstanceState(ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s in %s", errResourceNotFound, name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("%w: %s in %s", errNoPrimaryInstance, name, ms.Path)
	}

	return is, nil
}

// adpated from private function 'testCheckTypeSetElemNestedAttrsInState' from  github.com/hashicorp/terraform-plugin-testing/helper/resource
// testCheckTypeSetElemNestedAttrsInStateWithPair is a helper function
// to determine if nested attributes and their values are equal to those
// in the instance state. Currently, the function accepts a "values" param of type
// map[string]string or map[string]*regexp.Regexp.
// In case of map[string]string, this function can be used to match attribute from another resource (TestCheckResourceAttrPair behavior's)
// Returns nil if all attributes match, else an error.
func testCheckTypeSetElemNestedAttrsInStateWithPair(s *terraform.State, name, attr string, matchCount int, values interface{}) error {
	var errs []error
	matches := make(map[string]int)

	is, err := primaryInstanceState(s, name)
	if err != nil {
		return fmt.Errorf("failed to get primary instance state for %s: %w", name, err)
	}

	attrParts := strings.Split(attr, ".")
	if attrParts[len(attrParts)-1] != sentinelIndex {
		return fmt.Errorf("%w: %q does not end with the special value %q", errMissingSpecialSuffix, attr, sentinelIndex)
	}

	for stateKey, stateValue := range is.Attributes {
		stateKeyParts := strings.Split(stateKey, ".")
		if !isCandidateStateKey(stateKeyParts, attrParts) {
			continue
		}
		if !pathMatches(attrParts, stateKeyParts) {
			continue
		}

		id := stateKeyParts[len(attrParts)-1]
		nestedAttr := strings.Join(stateKeyParts[len(attrParts):], ".")

		match, err := matchNestedAttrValue(is, s, name, stateKey, stateValue, nestedAttr, values, &errs)
		if err != nil {
			return err
		}
		if match {
			matches[id]++
			if matches[id] == matchCount {
				return nil
			}
		}
	}

	return fmt.Errorf(
		"%w: %q no TypeSet element %q, with nested attrs %#v in state: %v",
		errTypeSetElementNotFound,
		name,
		attr,
		values,
		errors.Join(errs...),
	)
}

func isCandidateStateKey(stateKeyParts, attrParts []string) bool {
	// a Set/List item with nested attrs would have a flatmap address of
	// at least length 3, e.g. foo.0.name = "bar"
	return len(stateKeyParts) >= 3 && len(attrParts) <= len(stateKeyParts)
}

func pathMatches(attrParts, stateKeyParts []string) bool {
	for i := range attrParts {
		if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
			return false
		}
	}

	return true
}

func matchNestedAttrValue(
	isFirst *terraform.InstanceState,
	s *terraform.State,
	name, stateKey, stateValue, nestedAttr string,
	values interface{},
	errs *[]error,
) (bool, error) {
	switch t := values.(type) {
	case map[string]string:
		return matchStringNestedAttr(isFirst, s, name, stateKey, stateValue, nestedAttr, t, errs)
	case map[string]*regexp.Regexp:
		if v, keyExists := t[nestedAttr]; keyExists && v != nil && v.MatchString(stateValue) {
			return true, nil
		}
	}

	return false, nil
}

func matchStringNestedAttr(
	isFirst *terraform.InstanceState,
	s *terraform.State,
	name, stateKey, stateValue, nestedAttr string,
	values map[string]string,
	errs *[]error,
) (bool, error) {
	var match bool

	if strings.HasPrefix(values[nestedAttr], PairPrefix) {
		pairedValue := strings.TrimPrefix(values[nestedAttr], PairPrefix)
		if !strings.Contains(pairedValue, ".") {
			return false, fmt.Errorf(
				"%w: %s. Field starting with prefix %s must contain resource name and resource field separated by a dot",
				errInvalidPairFieldName,
				pairedValue,
				PairPrefix,
			)
		}

		lastDotIndex := strings.LastIndex(pairedValue, ".")
		nameSecond, attrSecond := pairedValue[:lastDotIndex], pairedValue[lastDotIndex+1:]
		isSecond, err := primaryInstanceState(s, nameSecond)
		if err != nil {
			return false, fmt.Errorf("failed to get primary instance state for %s: %w", nameSecond, err)
		}

		if err := testCheckResourceAttrPair(isFirst, name, stateKey, isSecond, nameSecond, attrSecond); err != nil {
			*errs = append(*errs, err)
		} else {
			match = true
		}
	}

	if v, keyExists := values[nestedAttr]; keyExists && v == stateValue {
		match = true
	}

	return match, nil
}
