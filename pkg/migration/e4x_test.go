package migration

import (
	"testing"

	"github.com/sroopra/ghega/pkg/mapping"
	"github.com/sroopra/ghega/pkg/mirthxml"
)

func TestClassifyTransformerStep_Empty(t *testing.T) {
	step := mirthxml.Step{Script: ""}
	res := ClassifyTransformerStep(step)
	if res.Status != "auto_converted" {
		t.Errorf("expected auto_converted, got %s", res.Status)
	}
	if len(res.Patterns) != 0 {
		t.Errorf("expected 0 patterns, got %d", len(res.Patterns))
	}
}

func TestClassifyTransformerStep_SimpleFieldAssignment_Static(t *testing.T) {
	step := mirthxml.Step{Script: "msg['PID']['PID.3']['PID.3.1'] = 'UNKNOWN';"}
	res := ClassifyTransformerStep(step)
	if len(res.Patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(res.Patterns))
	}
	p := res.Patterns[0]
	if p.Category != CategoryFieldAssignment {
		t.Errorf("expected field_assignment, got %s", p.Category)
	}
	if p.Disposition != DispositionAutoConvertible {
		t.Errorf("expected auto_convertible, got %s", p.Disposition)
	}
	if p.Mapping == nil {
		t.Fatalf("expected mapping")
	}
	if p.Mapping.Target != "PID-3.1" {
		t.Errorf("expected target PID-3.1, got %s", p.Mapping.Target)
	}
	if p.Mapping.Transform != mapping.TransformStatic {
		t.Errorf("expected static transform, got %s", p.Mapping.Transform)
	}
	if p.Mapping.Value != "UNKNOWN" {
		t.Errorf("expected value UNKNOWN, got %s", p.Mapping.Value)
	}
	if res.Status != "auto_converted" {
		t.Errorf("expected auto_converted status, got %s", res.Status)
	}
}

func TestClassifyTransformerStep_SimpleFieldAssignment_Copy(t *testing.T) {
	step := mirthxml.Step{Script: "msg['PID']['PID.3']['PID.3.1'] = msg['PID']['PID.5']['PID.5.1'];"}
	res := ClassifyTransformerStep(step)
	if len(res.Patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(res.Patterns))
	}
	p := res.Patterns[0]
	if p.Disposition != DispositionAutoConvertible {
		t.Errorf("expected auto_convertible, got %s", p.Disposition)
	}
	if p.Mapping == nil {
		t.Fatalf("expected mapping")
	}
	if p.Mapping.Source != "PID-5.1" {
		t.Errorf("expected source PID-5.1, got %s", p.Mapping.Source)
	}
	if p.Mapping.Target != "PID-3.1" {
		t.Errorf("expected target PID-3.1, got %s", p.Mapping.Target)
	}
	if p.Mapping.Transform != mapping.TransformCopy {
		t.Errorf("expected copy transform, got %s", p.Mapping.Transform)
	}
}

func TestClassifyTransformerStep_Conditional(t *testing.T) {
	step := mirthxml.Step{Script: "if (msg['PID']['PID.3']['PID.3.1'] == '') { msg['PID']['PID.3']['PID.3.1'] = 'UNKNOWN'; }"}
	res := ClassifyTransformerStep(step)
	if len(res.Patterns) < 1 {
		t.Fatalf("expected at least 1 pattern, got %d", len(res.Patterns))
	}
	foundCond := false
	for _, p := range res.Patterns {
		if p.Category == CategoryConditional {
			foundCond = true
			if p.Disposition != DispositionNeedsRewrite {
				t.Errorf("expected needs_rewrite for conditional, got %s", p.Disposition)
			}
		}
	}
	if !foundCond {
		t.Errorf("expected conditional pattern")
	}
}

func TestClassifyTransformerStep_Loop(t *testing.T) {
	step := mirthxml.Step{Script: "for each (seg in msg.children()) { logger.info(seg); }"}
	res := ClassifyTransformerStep(step)
	foundLoop := false
	for _, p := range res.Patterns {
		if p.Category == CategoryLoop {
			foundLoop = true
			if p.Disposition != DispositionNeedsRewrite {
				t.Errorf("expected needs_rewrite for loop, got %s", p.Disposition)
			}
		}
	}
	if !foundLoop {
		t.Errorf("expected loop pattern")
	}
}

func TestClassifyTransformerStep_E4X(t *testing.T) {
	step := mirthxml.Step{Script: "var xml = new XML('<root/>'); var nodes = xml..*; var attr = xml.@id;"}
	res := ClassifyTransformerStep(step)
	foundE4X := false
	for _, p := range res.Patterns {
		if p.Category == CategoryE4XManipulation {
			foundE4X = true
			if p.Disposition != DispositionUnsupported {
				t.Errorf("expected unsupported for E4X, got %s", p.Disposition)
			}
		}
	}
	if !foundE4X {
		t.Errorf("expected E4X pattern")
	}
}

func TestClassifyTransformerStep_DestinationDispatch(t *testing.T) {
	step := mirthxml.Step{Script: "destinationSet.removeAll(); router.routeMessage(msg);"}
	res := ClassifyTransformerStep(step)
	foundDest := false
	for _, p := range res.Patterns {
		if p.Category == CategoryDestinationDispatch {
			foundDest = true
			if p.Disposition != DispositionNeedsRewrite {
				t.Errorf("expected needs_rewrite for destination dispatch, got %s", p.Disposition)
			}
		}
	}
	if !foundDest {
		t.Errorf("expected destination dispatch pattern")
	}
}

func TestClassifyTransformerStep_Logger(t *testing.T) {
	step := mirthxml.Step{Script: "logger.info('Processing message'); logger.debug(msg);"}
	res := ClassifyTransformerStep(step)
	foundLogger := false
	for _, p := range res.Patterns {
		if p.Category == CategoryLogger {
			foundLogger = true
			if p.Disposition != DispositionNeedsRewrite {
				t.Errorf("expected needs_rewrite for logger, got %s", p.Disposition)
			}
		}
	}
	if !foundLogger {
		t.Errorf("expected logger pattern")
	}
}

func TestClassifyTransformerStep_ExternalCall(t *testing.T) {
	step := mirthxml.Step{Script: "var result = myCustomLibrary.process(msg);"}
	res := ClassifyTransformerStep(step)
	foundExt := false
	for _, p := range res.Patterns {
		if p.Category == CategoryExternalCall {
			foundExt = true
			if p.Disposition != DispositionUnsupported {
				t.Errorf("expected unsupported for external call, got %s", p.Disposition)
			}
		}
	}
	if !foundExt {
		t.Errorf("expected external call pattern")
	}
}

func TestClassifyTransformerStep_ComplexRHS(t *testing.T) {
	step := mirthxml.Step{Script: "msg['PID']['PID.3']['PID.3.1'] = msg['PID']['PID.3']['PID.3.1'].toString().toUpperCase();"}
	res := ClassifyTransformerStep(step)
	if len(res.Patterns) != 1 {
		t.Fatalf("expected 1 pattern, got %d", len(res.Patterns))
	}
	p := res.Patterns[0]
	if p.Category != CategoryFieldAssignment {
		t.Errorf("expected field_assignment, got %s", p.Category)
	}
	if p.Disposition != DispositionNeedsRewrite {
		t.Errorf("expected needs_rewrite for complex RHS, got %s", p.Disposition)
	}
}

func TestClassifyTransformerStep_MixedStatus(t *testing.T) {
	step := mirthxml.Step{
		Script: `msg['PID']['PID.3']['PID.3.1'] = 'STATIC';
if (msg['PID']['PID.3']['PID.3.1'] == '') {
    msg['PID']['PID.3']['PID.3.1'] = 'UNKNOWN';
}`,
	}
	res := ClassifyTransformerStep(step)
	if res.Status != "mixed" {
		t.Errorf("expected mixed status, got %s", res.Status)
	}
}
