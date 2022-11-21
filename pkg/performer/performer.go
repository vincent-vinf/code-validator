package performer

import (
	"fmt"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
)

type Performer struct {
	id int
}

func New(id int) *Performer {
	return &Performer{
		id: id,
	}
}

func (p *Performer) Run(task *Task) (*Report, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	lang, err := NewLang(task.Runtime)
	if err != nil {
		return nil, err
	}

	verifyTemp, err := task.Verify.ToTemplate()
	if err != nil {
		return nil, err
	}
	templates := append(lang.GetTemplates(), *verifyTemp)

	// todo add init step
	steps := append(lang.GetRunSteps(),
		pipeline.Step{
			Name:           VerifyStepName,
			Template:       VerifyStepName,
			ContinueOnFail: true,
			FileRefs: []pipeline.FileRef{
				{
					DataRef: &pipeline.DataRef{
						ExternalRef: &pipeline.ExternalRef{FileName: "answer"},
					},
					Path:       "./answer",
					AutoRemove: true,
				},
				{
					DataRef: &pipeline.DataRef{
						StepOutRef: &pipeline.StepOutRef{StepName: RunStepName},
					},
					Path:       "./output",
					AutoRemove: true,
				},
			},
		},
	)

	rep := &Report{
		Result: "pass",
	}
	for _, tc := range task.Cases {
		pl := &pipeline.Pipeline{
			Name:      task.Name,
			Steps:     steps,
			Templates: templates,
			Files: []pipeline.File{
				{
					Name:    "code",
					Content: task.Run.SourceCode,
				},
				{
					Name:    "answer",
					Content: tc.Output,
				},
				{
					Name:    "input",
					Content: tc.Input,
				},
			},
		}
		executor, err := pipeline.NewExecutor(p.id)
		if err != nil {
			return nil, err
		}

		res, err := executor.Exec(*pl)
		if err != nil {
			return nil, fmt.Errorf("exec fail, err: %w", err)
		}

		if err = executor.Clean(); err != nil {
			return nil, err
		}

		casePass := true
		if _, ok := res.Errs[VerifyStepName]; ok {
			rep.Result = "fail"
			casePass = false
		}
		meta, ok := res.Metas[RunStepName]
		if !ok {
			return nil, fmt.Errorf("the metadata of test case %s is missing", tc.Name)
		}
		rep.Cases = append(rep.Cases, CaseResult{
			Name:   tc.Name,
			Result: casePass,

			ExitCode: meta.ExitCode,
			Time:     meta.Time,
			Memory:   meta.MaxRSS,
		})

	}

	return rep, nil
}
