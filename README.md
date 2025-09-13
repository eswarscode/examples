# Multi-Stage Pipeline with Failure Handling

A robust Go implementation of a multi-stage pipeline with comprehensive failure handling, automatic retries, and restart capabilities.

## Key Features

### Pipeline Structure
- `Stage` interface with configurable retries, timeouts, and dependencies
- `Pipeline` manager with concurrent execution support
- Thread-safe operations with proper mutex handling

### Failure Handling
- Automatic retry mechanism with configurable delays
- Per-stage retry limits and timeout settings
- Graceful failure handling with detailed error reporting

### Restart Functionality
- `RestartFailedStages()` - restart all failed stages
- `RestartStage(name)` - restart specific stage and its dependents
- `Reset()` - reset entire pipeline to initial state

### Execution Features
- Dependency-based stage ordering
- Concurrent execution with configurable limits
- Global and per-stage timeouts
- Fail-fast or continue-on-failure modes

## Usage Example

The example demonstrates a 4-stage data processing pipeline:
1. **Data Processing** → **Validation** → **Transformation** → **Output**

Each stage has different failure rates and retry configurations, showing how the pipeline handles failures and automatically retries failed stages.

## Quick Start

```bash
go run .
```

This will execute the example pipeline with realistic failure scenarios and recovery mechanisms.

## API Overview

### Creating a Pipeline

```go
config := PipelineConfig{
    MaxConcurrency:    3,
    FailFast:          false,
    ContinueOnFailure: true,
    GlobalTimeout:     time.Minute * 2,
}

pipeline := NewPipeline(config, logger)
```

### Implementing Custom Stages

```go
type MyStage struct {
    *BaseStage
}

func NewMyStage() *MyStage {
    return &MyStage{
        BaseStage: NewBaseStage("my_stage", []string{"dependency_stage"}).
            SetMaxRetries(3).
            SetRetryDelay(time.Second * 2).
            SetTimeout(time.Second * 30),
    }
}

func (s *MyStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    // Your stage logic here
    return result, nil
}
```

### Pipeline Management

```go
// Add stages
pipeline.AddStage(NewMyStage())

// Execute pipeline
err := pipeline.Execute()

// Check status
pipeline.PrintStatus()

// Restart failed stages
pipeline.RestartFailedStages()

// Restart specific stage
pipeline.RestartStage("stage_name")

// Reset pipeline
pipeline.Reset()
```

## Stage Configuration

Each stage can be configured with:
- **Dependencies**: Other stages that must complete first
- **Max Retries**: Number of retry attempts on failure
- **Retry Delay**: Time to wait between retries
- **Timeout**: Maximum execution time per attempt

## Pipeline Configuration

- **MaxConcurrency**: Maximum number of stages running simultaneously
- **FailFast**: Stop execution on first failure
- **ContinueOnFailure**: Continue executing independent stages after failures
- **GlobalTimeout**: Maximum total pipeline execution time

## Error Handling

The pipeline provides comprehensive error handling:
- Individual stage failures with retry logic
- Dependency validation
- Timeout handling
- Graceful shutdown on cancellation
- Detailed error reporting and logging

## Monitoring

Track pipeline progress with:
- Real-time status updates
- Execution duration tracking
- Attempt counters
- Error details
- Output capture

Run the example to see all these features in action!