package perform

import (
	"errors"
	"fmt"
	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/types"
	"github.com/vincent-vinf/code-validator/pkg/util/dispatcher"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
	"strings"
)

const (
	Runtime = types.PythonRuntime

	InitStepName   = "init"
	RunStepName    = "run"
	VerifyStepName = "verify"
)

var (
	idDispatcher *dispatcher.Dispatcher
	ossClient    *oss.Client
)

func init() {
	var err error
	idDispatcher, err = dispatcher.NewDispatcher(100, 500)
	if err != nil {
		panic(err)
	}
}

func Perform(vf *Verification, codePath string) (*Report, error) {
	if err := validate(vf); err != nil {
		return nil, err
	}
	switch {
	case vf.Code != nil:
		return runCode(vf.Code, codePath)
	case vf.Custom != nil:
		return runCustom(vf.Custom, codePath)
	default:
		return nil, errors.New("verification name cannot be empty")
	}
}

func runCode(code *CodeVerification, codePath string) (*Report, error) {
	var steps []pipeline.Step
	if code.Init != nil {
		code.Init.Name = InitStepName
		steps = append(steps, *code.Init.ToStep())
	}
	templates := GetCodeTemplates()
	steps = append(steps, GetCodeSteps()...)
	steps = append(steps, pipeline.Step{
		Name: VerifyStepName,
		InlineTemplate: &pipeline.Template{
			Name: VerifyStepName,
			Cmd:  "/bin/sh",
			Args: []string{
				"-c",
				code.Verify,
			},
		},
		ContinueOnFail: true,
		FileRefs: []pipeline.FileRef{
			{
				DataRef: &pipeline.DataRef{
					ExternalRef: &pipeline.ExternalRef{FileName: "output"},
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
		Pass:  true,
		Cases: nil,
	}
	codeFile := File{
		OssPath: codePath,
	}
	codeData, err := codeFile.Read()
	if err != nil {
		rep.Pass = false
		rep.Message = fmt.Sprintf("failed to get code file, err: %s", err)

		return rep, nil
	}
	id, err := idDispatcher.Get()
	if err != nil {
		// Too many verification items are running at the same time,
		// an error is returned, and the upper layer will retry
		return nil, fmt.Errorf("too many validations running at the same time: %w", err)
	}
	defer idDispatcher.Release(id)

	for _, tc := range code.Cases {
		cr := CaseResult{
			Name: tc.Name,
			Pass: false,
		}

		inData, err := tc.In.Read()
		if err != nil {
			// Failed to read file, skip test case
			cr.Message = err.Error()
			rep.Cases = append(rep.Cases, cr)
			continue
		}
		outData, err := tc.Out.Read()
		if err != nil {
			cr.Message = err.Error()
			rep.Cases = append(rep.Cases, cr)
			continue
		}
		pl := &pipeline.Pipeline{
			Steps:     steps,
			Templates: templates,
			Files: []pipeline.File{
				{
					Name:    "code",
					Content: codeData,
				},
				{
					Name:    "output",
					Content: outData,
				},
				{
					Name:    "input",
					Content: inData,
				},
			},
		}
		res, err := execute(id, pl)
		if err != nil {
			return nil, err
		}
		if _, ok := res.Errs[VerifyStepName]; ok {
			cr.Pass = false
			rep.Pass = false
		} else {
			cr.Pass = true
		}
		meta, ok := res.Metas[RunStepName]
		if !ok {
			cr.Message = fmt.Sprintf("the metadata of test case %s is missing", tc.Name)
		} else {
			cr.ExitCode = meta.ExitCode
			cr.Time = meta.Time
			cr.Memory = meta.MaxRSS
		}
		rep.Cases = append(rep.Cases, cr)
	}

	return rep, nil
}

func runCustom(custom *CustomVerification, codePath string) (*Report, error) {
	rep := &Report{
		Pass: true,
	}
	codeFile := File{
		OssPath: codePath,
	}
	codeData, err := codeFile.Read()
	if err != nil {
		rep.Pass = false
		rep.Message = fmt.Sprintf("failed to get code file, err: %s", err)

		return rep, nil
	}

	id, err := idDispatcher.Get()
	if err != nil {
		// Too many verification items are running at the same time,
		// an error is returned, and the upper layer will retry
		return nil, fmt.Errorf("too many validations running at the same time: %w", err)
	}
	defer idDispatcher.Release(id)

	var steps []pipeline.Step
	for _, action := range custom.Actions {
		steps = append(steps, *action.ToStep())
	}
	pl := &pipeline.Pipeline{
		Steps: steps,
		Files: []pipeline.File{
			{
				Name:    "code",
				Content: codeData,
			},
		},
	}

	res, err := execute(id, pl)
	if err != nil {
		return nil, err
	}
	if len(res.Errs) > 0 {
		rep.Pass = false
		var msgs []string
		for step, e := range res.Errs {
			msgs = append(msgs, fmt.Sprintf("step %s error: %s.", step, e))
		}
		rep.Message = strings.Join(msgs, "\n")

		return rep, nil
	}
	rep.Pass, rep.Message, err = readVerifyResult()
	if err != nil {
		return nil, err
	}

	return rep, nil
}

func readVerifyResult() (pass bool, message string, err error) {
	return true, "test msg", nil
}

func validate(vf *Verification) error {
	if vf == nil {
		return errors.New("verification cannot be empty")
	}
	if vf.Runtime != Runtime {
		return errors.New("runtime mismatch")
	}
	if vf.Name == "" {
		return errors.New("verification name cannot be empty")
	}

	return nil
}

func SetOssClient(c *oss.Client) {
	ossClient = c
}

func execute(id int, pl *pipeline.Pipeline) (res *pipeline.Result, err error) {
	executor, err := pipeline.NewExecutor(id)
	if err != nil {
		return nil, err
	}
	//defer func(executor *pipeline.Executor) {
	//	if e := executor.Clean(); e != nil {
	//		err = e
	//	}
	//}(executor)
	res, err = executor.Exec(*pl)

	return
}
