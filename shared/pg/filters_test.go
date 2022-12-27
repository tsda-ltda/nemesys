package pg

import "testing"

func TestApplyFilters(t *testing.T) {
	tests := []struct {
		expect  string
		filters BasicContainersQueryFilters
	}{
		{
			expect: "",
			filters: BasicContainersQueryFilters{
				CreatedAtStart: 111999,
			},
		},
		{
			expect: "WHERE name ILIKE 'value%'",
			filters: BasicContainersQueryFilters{
				Name: "value",
			},
		},
		{
			expect: " ORDER BY name ASC",
			filters: BasicContainersQueryFilters{
				OrderBy:   "name",
				OrderByFn: "ASC",
			},
		},
	}

	for _, test := range tests {
		result, err := applyFilters(test.filters, "%s", BasicContainerValidOrderByColumns)
		if err != nil {
			t.Error(err)
			return
		}
		if result != test.expect {
			t.Errorf("expected: %s, got: %s", test.expect, result)
			return
		}
	}
}
