package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

type DataProcessingStage struct {
	*BaseStage
}

func NewDataProcessingStage() *DataProcessingStage {
	return &DataProcessingStage{
		BaseStage: NewBaseStage("data_processing", []string{}).
			SetMaxRetries(2).
			SetRetryDelay(time.Second * 1).
			SetTimeout(time.Second * 10),
	}
}

func (s *DataProcessingStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * 500)
	
	if rand.Float32() < 0.3 {
		return nil, errors.New("random data processing failure")
	}
	
	data := map[string]interface{}{
		"processed_records": 1000,
		"timestamp":         time.Now(),
		"source":           "data_processing",
	}
	
	return data, nil
}

type ValidationStage struct {
	*BaseStage
}

func NewValidationStage() *ValidationStage {
	return &ValidationStage{
		BaseStage: NewBaseStage("validation", []string{"data_processing"}).
			SetMaxRetries(1).
			SetRetryDelay(time.Second * 2).
			SetTimeout(time.Second * 5),
	}
}

func (s *ValidationStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * 300)
	
	if input == nil {
		return nil, errors.New("no input data to validate")
	}
	
	data, ok := input.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid input format")
	}
	
	if rand.Float32() < 0.2 {
		return nil, errors.New("validation failed")
	}
	
	result := map[string]interface{}{
		"validation_passed": true,
		"input_records":     data["processed_records"],
		"validated_at":      time.Now(),
	}
	
	return result, nil
}

type TransformationStage struct {
	*BaseStage
}

func NewTransformationStage() *TransformationStage {
	return &TransformationStage{
		BaseStage: NewBaseStage("transformation", []string{"validation"}).
			SetMaxRetries(3).
			SetRetryDelay(time.Second * 1).
			SetTimeout(time.Second * 8),
	}
}

func (s *TransformationStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * 700)
	
	if rand.Float32() < 0.25 {
		return nil, errors.New("transformation error")
	}
	
	result := map[string]interface{}{
		"transformed_data": "processed and validated data",
		"transformation_id": fmt.Sprintf("tx_%d", time.Now().UnixNano()),
		"completed_at":      time.Now(),
	}
	
	return result, nil
}

type OutputStage struct {
	*BaseStage
}

func NewOutputStage() *OutputStage {
	return &OutputStage{
		BaseStage: NewBaseStage("output", []string{"transformation"}).
			SetMaxRetries(2).
			SetRetryDelay(time.Second * 2).
			SetTimeout(time.Second * 5),
	}
}

func (s *OutputStage) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	time.Sleep(time.Millisecond * 200)
	
	if rand.Float32() < 0.15 {
		return nil, errors.New("output stage failure")
	}
	
	result := map[string]interface{}{
		"output_written": true,
		"output_size":    "1.2MB",
		"output_path":    "/tmp/pipeline_output.json",
	}
	
	return result, nil
}

func main() {
	logger := log.New(os.Stdout, "[PIPELINE] ", log.LstdFlags)
	
	config := PipelineConfig{
		MaxConcurrency:    3,
		FailFast:          false,
		ContinueOnFailure: true,
		GlobalTimeout:     time.Minute * 2,
	}
	
	pipeline := NewPipeline(config, logger)
	
	pipeline.AddStage(NewDataProcessingStage())
	pipeline.AddStage(NewValidationStage())
	pipeline.AddStage(NewTransformationStage())
	pipeline.AddStage(NewOutputStage())
	
	fmt.Println("=== Starting Pipeline Execution ===")
	
	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Printf("\n--- Execution Attempt %d ---\n", attempt)
		
		err := pipeline.Execute()
		pipeline.PrintStatus()
		
		if err != nil {
			fmt.Printf("Pipeline execution failed: %v\n", err)
		}
		
		if !pipeline.hasFailures() {
			fmt.Println("\n✅ Pipeline completed successfully!")
			break
		}
		
		if attempt < 3 {
			fmt.Println("\n🔄 Restarting failed stages...")
			pipeline.RestartFailedStages()
			time.Sleep(time.Second * 2)
		} else {
			fmt.Println("\n❌ Pipeline failed after 3 attempts")
		}
	}
	
	fmt.Println("\n=== Final Results ===")
	results := pipeline.GetAllResults()
	for name, result := range results {
		fmt.Printf("Stage %s: %s", name, result.Status)
		if result.Error != nil {
			fmt.Printf(" (Error: %v)", result.Error)
		}
		fmt.Println()
	}
	
	fmt.Println("\n=== Demonstrating Manual Restart ===")
	fmt.Println("Manually restarting 'validation' stage...")
	pipeline.RestartStage("validation")
	pipeline.Execute()
	pipeline.PrintStatus()
}