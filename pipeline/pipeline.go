package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type StageStatus int

const (
	StatusPending StageStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusSkipped
)

func (s StageStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusRunning:
		return "RUNNING"
	case StatusCompleted:
		return "COMPLETED"
	case StatusFailed:
		return "FAILED"
	case StatusSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

type StageResult struct {
	Status    StageStatus
	Error     error
	Output    interface{}
	Duration  time.Duration
	Attempts  int
	StartTime time.Time
	EndTime   time.Time
}

type Stage interface {
	Name() string
	Execute(ctx context.Context, input interface{}) (interface{}, error)
	Dependencies() []string
	MaxRetries() int
	RetryDelay() time.Duration
	Timeout() time.Duration
}

type BaseStage struct {
	name         string
	dependencies []string
	maxRetries   int
	retryDelay   time.Duration
	timeout      time.Duration
}

func NewBaseStage(name string, deps []string) *BaseStage {
	return &BaseStage{
		name:         name,
		dependencies: deps,
		maxRetries:   3,
		retryDelay:   time.Second * 2,
		timeout:      time.Minute * 5,
	}
}

func (s *BaseStage) Name() string              { return s.name }
func (s *BaseStage) Dependencies() []string    { return s.dependencies }
func (s *BaseStage) MaxRetries() int           { return s.maxRetries }
func (s *BaseStage) RetryDelay() time.Duration { return s.retryDelay }
func (s *BaseStage) Timeout() time.Duration    { return s.timeout }

func (s *BaseStage) SetMaxRetries(retries int) *BaseStage {
	s.maxRetries = retries
	return s
}

func (s *BaseStage) SetRetryDelay(delay time.Duration) *BaseStage {
	s.retryDelay = delay
	return s
}

func (s *BaseStage) SetTimeout(timeout time.Duration) *BaseStage {
	s.timeout = timeout
	return s
}

type PipelineConfig struct {
	MaxConcurrency    int
	FailFast          bool
	ContinueOnFailure bool
	GlobalTimeout     time.Duration
}

type Pipeline struct {
	stages   map[string]Stage
	results  map[string]*StageResult
	config   PipelineConfig
	mu       sync.RWMutex
	logger   *log.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewPipeline(config PipelineConfig, logger *log.Logger) *Pipeline {
	if logger == nil {
		logger = log.Default()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Pipeline{
		stages:  make(map[string]Stage),
		results: make(map[string]*StageResult),
		config:  config,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (p *Pipeline) AddStage(stage Stage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.stages[stage.Name()] = stage
	p.results[stage.Name()] = &StageResult{
		Status: StatusPending,
	}
}

func (p *Pipeline) GetStageResult(name string) (*StageResult, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	result, exists := p.results[name]
	return result, exists
}

func (p *Pipeline) GetAllResults() map[string]*StageResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	results := make(map[string]*StageResult)
	for name, result := range p.results {
		results[name] = result
	}
	return results
}

func (p *Pipeline) validateDependencies() error {
	for _, stage := range p.stages {
		for _, dep := range stage.Dependencies() {
			if _, exists := p.stages[dep]; !exists {
				return fmt.Errorf("stage %s depends on non-existent stage %s", stage.Name(), dep)
			}
		}
	}
	return nil
}

func (p *Pipeline) getExecutableStages() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	var executable []string
	for name, stage := range p.stages {
		result := p.results[name]
		if result.Status != StatusPending {
			continue
		}
		
		canExecute := true
		for _, dep := range stage.Dependencies() {
			depResult := p.results[dep]
			if depResult.Status != StatusCompleted {
				canExecute = false
				break
			}
		}
		
		if canExecute {
			executable = append(executable, name)
		}
	}
	return executable
}

func (p *Pipeline) executeStageWithRetry(stage Stage, input interface{}) {
	name := stage.Name()
	maxRetries := stage.MaxRetries()
	
	for attempt := 1; attempt <= maxRetries+1; attempt++ {
		p.logger.Printf("Starting execution of stage: %s (attempt %d/%d)", name, attempt, maxRetries+1)
		
		p.mu.Lock()
		result := p.results[name]
		result.Status = StatusRunning
		result.StartTime = time.Now()
		result.Attempts = attempt
		p.mu.Unlock()
		
		ctx, cancel := context.WithTimeout(p.ctx, stage.Timeout())
		output, err := stage.Execute(ctx, input)
		cancel()
		
		p.mu.Lock()
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Output = output
		result.Error = err
		
		if err == nil {
			result.Status = StatusCompleted
			p.logger.Printf("Stage %s completed successfully on attempt %d", name, attempt)
			p.mu.Unlock()
			return
		}
		
		result.Status = StatusFailed
		p.logger.Printf("Stage %s failed on attempt %d: %v", name, attempt, err)
		
		if attempt <= maxRetries {
			p.logger.Printf("Retrying stage %s in %v", name, stage.RetryDelay())
			p.mu.Unlock()
			
			select {
			case <-time.After(stage.RetryDelay()):
			case <-p.ctx.Done():
				p.mu.Lock()
				result.Error = fmt.Errorf("retry cancelled: %w", p.ctx.Err())
				p.mu.Unlock()
				return
			}
		} else {
			p.logger.Printf("Stage %s failed permanently after %d attempts", name, attempt)
			p.mu.Unlock()
		}
	}
}

func (p *Pipeline) Execute() error {
	if err := p.validateDependencies(); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}
	
	p.logger.Println("Starting pipeline execution")
	
	if p.config.GlobalTimeout > 0 {
		var cancel context.CancelFunc
		p.ctx, cancel = context.WithTimeout(p.ctx, p.config.GlobalTimeout)
		defer cancel()
	}
	
	semaphore := make(chan struct{}, p.config.MaxConcurrency)
	var wg sync.WaitGroup
	
	for {
		select {
		case <-p.ctx.Done():
			return fmt.Errorf("pipeline execution cancelled: %w", p.ctx.Err())
		default:
		}
		
		executable := p.getExecutableStages()
		if len(executable) == 0 {
			break
		}
		
		for _, stageName := range executable {
			stage := p.stages[stageName]
			
			wg.Add(1)
			go func(s Stage) {
				defer wg.Done()
				
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				var input interface{}
				if len(s.Dependencies()) > 0 {
					dep := s.Dependencies()[0]
					if depResult, exists := p.results[dep]; exists {
						input = depResult.Output
					}
				}
				
				p.executeStageWithRetry(s, input)
			}(stage)
		}
		
		wg.Wait()
		
		if p.config.FailFast && p.hasFailures() {
			return fmt.Errorf("pipeline execution stopped due to failures (fail-fast mode)")
		}
		
		if p.isComplete() {
			break
		}
	}
	
	p.logger.Println("Pipeline execution completed")
	return nil
}

func (p *Pipeline) hasFailures() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	for _, result := range p.results {
		if result.Status == StatusFailed {
			return true
		}
	}
	return false
}

func (p *Pipeline) isComplete() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	for _, result := range p.results {
		if result.Status == StatusPending || result.Status == StatusRunning {
			return false
		}
	}
	return true
}

func (p *Pipeline) RestartFailedStages() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	restarted := 0
	for name, result := range p.results {
		if result.Status == StatusFailed {
			p.logger.Printf("Restarting failed stage: %s", name)
			result.Status = StatusPending
			result.Error = nil
			result.Output = nil
			result.Attempts = 0
			result.StartTime = time.Time{}
			result.EndTime = time.Time{}
			result.Duration = 0
			restarted++
		}
	}
	
	p.logger.Printf("Restarted %d failed stages", restarted)
	return nil
}

func (p *Pipeline) RestartStage(stageName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result, exists := p.results[stageName]
	if !exists {
		return fmt.Errorf("stage %s not found", stageName)
	}
	
	p.logger.Printf("Restarting stage: %s", stageName)
	result.Status = StatusPending
	result.Error = nil
	result.Output = nil
	result.Attempts = 0
	result.StartTime = time.Time{}
	result.EndTime = time.Time{}
	result.Duration = 0
	
	dependentStages := p.getDependentStages(stageName)
	for _, depStage := range dependentStages {
		p.logger.Printf("Restarting dependent stage: %s", depStage)
		depResult := p.results[depStage]
		depResult.Status = StatusPending
		depResult.Error = nil
		depResult.Output = nil
		depResult.Attempts = 0
		depResult.StartTime = time.Time{}
		depResult.EndTime = time.Time{}
		depResult.Duration = 0
	}
	
	return nil
}

func (p *Pipeline) getDependentStages(stageName string) []string {
	var dependents []string
	for name, stage := range p.stages {
		for _, dep := range stage.Dependencies() {
			if dep == stageName {
				dependents = append(dependents, name)
				dependents = append(dependents, p.getDependentStages(name)...)
			}
		}
	}
	return dependents
}

func (p *Pipeline) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.logger.Println("Resetting pipeline")
	for name := range p.results {
		p.results[name] = &StageResult{
			Status: StatusPending,
		}
	}
}

func (p *Pipeline) Stop() {
	p.logger.Println("Stopping pipeline")
	p.cancel()
}

func (p *Pipeline) PrintStatus() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	fmt.Println("\n=== Pipeline Status ===")
	for name, result := range p.results {
		fmt.Printf("Stage: %-20s Status: %-10s Attempts: %d", name, result.Status, result.Attempts)
		if result.Duration > 0 {
			fmt.Printf(" Duration: %v", result.Duration)
		}
		if result.Error != nil {
			fmt.Printf(" Error: %v", result.Error)
		}
		fmt.Println()
	}
	fmt.Println("=====================")
}