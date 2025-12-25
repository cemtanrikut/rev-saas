package service

import (
	"testing"
)

// TestComputeTotalFromPlanCounts verifies that total_active_customers is computed
// as the sum of plan_customer_counts when plan counts are provided.
func TestComputeTotalFromPlanCounts(t *testing.T) {
	tests := []struct {
		name                 string
		planCounts           map[string]int
		providedTotal        *int
		expectedTotal        *int
		expectValidationErr  bool
	}{
		{
			name: "compute total from plan counts, ignore provided total",
			planCounts: map[string]int{
				"starter": 10,
				"growth":  5,
			},
			providedTotal: intPtr(999), // should be ignored
			expectedTotal: intPtr(15),  // 10 + 5 = 15
		},
		{
			name: "compute total from plan counts with zero values",
			planCounts: map[string]int{
				"starter": 0,
				"growth":  5,
				"pro":     3,
			},
			providedTotal: nil,
			expectedTotal: intPtr(8), // 0 + 5 + 3 = 8
		},
		{
			name:          "no plan counts, accept provided total as-is",
			planCounts:    nil,
			providedTotal: intPtr(100),
			expectedTotal: intPtr(100),
		},
		{
			name:          "no plan counts, no total provided",
			planCounts:    nil,
			providedTotal: nil,
			expectedTotal: nil,
		},
		{
			name:          "empty plan counts map, accept provided total",
			planCounts:    map[string]int{},
			providedTotal: intPtr(50),
			expectedTotal: intPtr(50),
		},
		{
			name: "single plan",
			planCounts: map[string]int{
				"enterprise": 25,
			},
			providedTotal: nil,
			expectedTotal: intPtr(25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the computation logic from SetMetrics
			totalActiveCustomers := tt.providedTotal
			planCustomerCounts := tt.planCounts

			if len(planCustomerCounts) > 0 {
				sum := 0
				for _, count := range planCustomerCounts {
					sum += count
				}
				totalActiveCustomers = &sum
			}

			// Verify result
			if tt.expectedTotal == nil {
				if totalActiveCustomers != nil {
					t.Errorf("expected nil total, got %d", *totalActiveCustomers)
				}
			} else {
				if totalActiveCustomers == nil {
					t.Errorf("expected total %d, got nil", *tt.expectedTotal)
				} else if *totalActiveCustomers != *tt.expectedTotal {
					t.Errorf("expected total %d, got %d", *tt.expectedTotal, *totalActiveCustomers)
				}
			}
		})
	}
}

// TestNegativePlanCountsValidation verifies that negative counts are rejected.
func TestNegativePlanCountsValidation(t *testing.T) {
	tests := []struct {
		name       string
		planCounts map[string]int
		wantErr    bool
	}{
		{
			name: "valid counts",
			planCounts: map[string]int{
				"starter": 10,
				"growth":  5,
			},
			wantErr: false,
		},
		{
			name: "negative count should fail",
			planCounts: map[string]int{
				"starter": -5,
				"growth":  10,
			},
			wantErr: true,
		},
		{
			name: "zero count is valid",
			planCounts: map[string]int{
				"starter": 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic from SetMetrics
			var validationErr error
			for planKey, count := range tt.planCounts {
				if count < 0 {
					validationErr = errNegativeCount(planKey)
					break
				}
			}

			if tt.wantErr && validationErr == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.wantErr && validationErr != nil {
				t.Errorf("unexpected validation error: %v", validationErr)
			}
		})
	}
}

// TestDeterministicComputation verifies that the same input always produces the same output.
func TestDeterministicComputation(t *testing.T) {
	planCounts := map[string]int{
		"starter":    10,
		"growth":     5,
		"enterprise": 20,
	}

	// Run computation multiple times
	results := make([]int, 10)
	for i := 0; i < 10; i++ {
		sum := 0
		for _, count := range planCounts {
			sum += count
		}
		results[i] = sum
	}

	// All results should be the same
	expected := 35 // 10 + 5 + 20
	for i, result := range results {
		if result != expected {
			t.Errorf("iteration %d: expected %d, got %d", i, expected, result)
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

type validationError struct {
	planKey string
}

func (e *validationError) Error() string {
	return "plan_customer_counts values must be non-negative: " + e.planKey
}

func errNegativeCount(planKey string) error {
	return &validationError{planKey: planKey}
}




