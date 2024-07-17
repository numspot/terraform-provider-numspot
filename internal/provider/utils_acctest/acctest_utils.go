package utils_acctest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	BASE_SUFFIX   = ""
	NEW_SUFFIX    = "new"
	PAIR_PREFIX   = "[[FROM PAIR]]"
	sentinelIndex = "*"
)

func ListToStringList(list []string) string {
	var strList string
	if len(list) > 0 {
		strList = fmt.Sprintf(`["%s"]`, strings.Join(list, `", "`))
	} else {
		strList = `[]`
	}

	return strList
}

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
			return fmt.Errorf("%#v has no non-empty values", values)
		}

		return testCheckTypeSetElemNestedAttrsInStateWithPair(s, name, attr, matchCount, values)
	}
}

// private function forked and unchanged from github.com/hashicorp/terraform-plugin-testing/helper/resource
func testCheckResourceAttrPair(isFirst *terraform.InstanceState, nameFirst string, keyFirst string, isSecond *terraform.InstanceState, nameSecond string, keySecond string) error {
	if nameFirst == nameSecond && keyFirst == keySecond {
		return fmt.Errorf(
			"comparing self: resource %s attribute %s",
			nameFirst,
			keyFirst,
		)
	}

	vFirst, okFirst := isFirst.Attributes[keyFirst]
	vSecond, okSecond := isSecond.Attributes[keySecond]

	// Container count values of 0 should not be relied upon, and not reliably
	// maintained by helper/schema. For the purpose of tests, consider unset and
	// 0 to be equal.
	if len(keyFirst) > 2 && len(keySecond) > 2 && keyFirst[len(keyFirst)-2:] == keySecond[len(keySecond)-2:] &&
		(strings.HasSuffix(keyFirst, ".#") || strings.HasSuffix(keyFirst, ".%")) {
		// they have the same suffix, and it is a collection count key.
		if vFirst == "0" || vFirst == "" {
			okFirst = false
		}
		if vSecond == "0" || vSecond == "" {
			okSecond = false
		}
	}

	if okFirst != okSecond {
		if !okFirst {
			return fmt.Errorf("%s: Attribute %q not set, but %q is set in %s as %q", nameFirst, keyFirst, keySecond, nameSecond, vSecond)
		}
		return fmt.Errorf("%s: Attribute %q is %q, but %q is not set in %s", nameFirst, keyFirst, vFirst, keySecond, nameSecond)
	}
	if !(okFirst || okSecond) {
		// If they both don't exist then they are equally unset, so that's okay.
		return nil
	}

	if vFirst != vSecond {
		return fmt.Errorf(
			"%s: Attribute '%s' expected %#v, got %#v",
			nameFirst,
			keyFirst,
			vSecond,
			vFirst)
	}

	return nil
}

// private function forked and unchanged from github.com/hashicorp/terraform-plugin-testing/helper/resource
// primaryInstanceState returns the primary instance state for the given
// resource name in the root module.
func primaryInstanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule() //nolint:staticcheck // legacy usage
	return modulePrimaryInstanceState(ms, name)
}

// private function forked and unchanged from github.com/hashicorp/terraform-plugin-testing/helper/resource
// modulePrimaryInstanceState returns the instance state for the given resource
// name in a ModuleState
func modulePrimaryInstanceState(ms *terraform.ModuleState, name string) (*terraform.InstanceState, error) {
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
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
func testCheckTypeSetElemNestedAttrsInStateWithPair(s *terraform.State, name string, attr string, matchCount int, values interface{}) error {
	matches := make(map[string]int)

	is, err := primaryInstanceState(s, name)
	if err != nil {
		return err
	}

	attrParts := strings.Split(attr, ".")
	if attrParts[len(attrParts)-1] != sentinelIndex {
		return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
	}

	for stateKey, stateValue := range is.Attributes {
		stateKeyParts := strings.Split(stateKey, ".")
		// a Set/List item with nested attrs would have a flatmap address of
		// at least length 3
		// foo.0.name = "bar"
		if len(stateKeyParts) < 3 || len(attrParts) > len(stateKeyParts) {
			continue
		}
		var pathMatch bool
		for i := range attrParts {
			if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
				break
			}
			if i == len(attrParts)-1 {
				pathMatch = true
			}
		}
		if !pathMatch {
			continue
		}
		id := stateKeyParts[len(attrParts)-1]
		nestedAttr := strings.Join(stateKeyParts[len(attrParts):], ".")

		var match bool
		switch t := values.(type) {
		case map[string]string:
			if strings.HasPrefix(t[nestedAttr], PAIR_PREFIX) {
				nameAttrSecond := strings.TrimPrefix(t[nestedAttr], PAIR_PREFIX)
				if !strings.Contains(nameAttrSecond, ".") {
					return fmt.Errorf("Invalid field name %s. Field starting with prefix %s must contain resource name and resource field separated by a dot. ", nameAttrSecond, PAIR_PREFIX)
				}
				lastDotIndex := strings.LastIndex(nameAttrSecond, ".")
				nameSecond, attrSec := nameAttrSecond[:lastDotIndex], nameAttrSecond[lastDotIndex+1:]
				isSecond, err := primaryInstanceState(s, nameSecond)
				if err != nil {
					return err
				}
				err = testCheckResourceAttrPair(is, name, stateKey, isSecond, nameSecond, attrSec)
				if err != nil {
					return err
				} else {
					match = true
				}
			}
			if v, keyExists := t[nestedAttr]; keyExists && v == stateValue {
				match = true
			}
		case map[string]*regexp.Regexp:
			if v, keyExists := t[nestedAttr]; keyExists && v != nil && v.MatchString(stateValue) {
				match = true
			}
		}
		if match {
			matches[id] = matches[id] + 1
			if matches[id] == matchCount {
				return nil
			}
		}
	}

	return fmt.Errorf("%q no TypeSet element %q, with nested attrs %#v in state", name, attr, values)
}
