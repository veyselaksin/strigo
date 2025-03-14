# Test Writing Best Practices

## General Guidelines

### 1. Test Organization
- Group related tests using subtests
- Use descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)

### 2. Test Independence
```go
func TestIndependent(t *testing.T) {
    // Setup
    cleanup := setUp()
    defer cleanup()

    // Test implementation
}
```

### 3. Error Handling
```go
func TestErrorCases(t *testing.T) {
    t.Run("Invalid Input", func(t *testing.T) {
        _, err := processInput("")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid input")
    })
}
```

## Common Patterns

### 1. Table-Driven Tests
```go
func TestTableDriven(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "TEST",
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: "",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := processInput(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. Setup and Teardown
```go
func TestWithSetup(t *testing.T) {
    // Setup
    rdb := helpers.NewRedisClient()
    defer helpers.CleanupRedis(t, rdb)

    // Test cases
    t.Run("first test", func(t *testing.T) {
        // Test implementation
    })

    t.Run("second test", func(t *testing.T) {
        // Test implementation
    })
}
``` 