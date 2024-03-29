package perform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/vincent-vinf/code-validator/pkg/pipeline"
	"github.com/vincent-vinf/code-validator/pkg/util/dispatcher"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
)

const (
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

func Perform(vf *Verification, codeOssPath string, srcDir, stepOutDir string) (*Report, error) {
	if err := validate(vf); err != nil {
		return nil, err
	}
	switch {
	case vf.Code != nil:
		return runCode(vf.Code, codeOssPath, srcDir, stepOutDir)
	case vf.Custom != nil:
		return runCustom(vf.Custom, codeOssPath, srcDir, stepOutDir)
	default:
		return nil, errors.New("verification name cannot be empty")
	}
}

func runCode(code *CodeVerification, codePath string, srcDir, stepOutDir string) (*Report, error) {
	var steps []pipeline.Step
	var files []pipeline.File
	if code.Init != nil {
		code.Init.Name = InitStepName
		steps = append(steps, *code.Init.ToStep())
		fs, err := code.Init.GetFiles(srcDir)
		if err != nil {
			return nil, err
		}
		files = append(files, fs...)
	}

	templates := GetCodeTemplates()
	steps = append(steps, GetCodeSteps()...)
	// get verify files
	fs, err := ToPipelineFile(srcDir, VerifyStepName, code.Files)
	if err != nil {
		return nil, err
	}
	files = append(files, fs...)
	var verifyFileRefs []pipeline.FileRef
	for i := range code.Files {
		verifyFileRefs = append(verifyFileRefs, pipeline.FileRef{
			DataRef: pipeline.DataRef{
				ExternalRef: &pipeline.ExternalRef{
					FileName: GetFileName(VerifyStepName, code.Files[i].Path),
				},
			},
			Path: code.Files[i].Path,
		})
	}
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
		FileRefs: append([]pipeline.FileRef{
			{
				DataRef: pipeline.DataRef{
					ExternalRef: &pipeline.ExternalRef{FileName: "output"},
				},
				Path:       "./answer",
				AutoRemove: true,
			},
			{
				DataRef: pipeline.DataRef{
					StepOutRef: &pipeline.StepOutRef{StepName: RunStepName},
				},
				Path:       "./output",
				AutoRemove: true,
			},
		}, verifyFileRefs...),
	},
	)

	rep := &Report{
		Pass:  true,
		Cases: nil,
	}

	codeData, err := ReadOSSFile(codePath)
	if err != nil {
		rep.Pass = false
		rep.Message = fmt.Sprintf("failed to get code file, path: %s, err: %s", codePath, err)

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

		inData, err := ReadOSSFile(path.Join(srcDir, tc.In.OssPath))
		if err != nil {
			// Failed to read file, skip test case
			cr.Message = err.Error()
			rep.Cases = append(rep.Cases, cr)
			continue
		}
		outData, err := ReadOSSFile(path.Join(srcDir, tc.Out.OssPath))
		if err != nil {
			cr.Message = err.Error()
			rep.Cases = append(rep.Cases, cr)
			continue
		}
		pl := &pipeline.Pipeline{
			Steps:     steps,
			Templates: templates,
			Files: append(files,
				pipeline.File{
					Name:    "code",
					Content: codeData,
				},
				pipeline.File{
					Name:    "output",
					Content: outData,
				},
				pipeline.File{
					Name:    "input",
					Content: inData,
				},
			),
		}
		res, _, _, err := execute(id, pl, path.Join(stepOutDir, tc.Name))
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

func runCustom(custom *CustomVerification, codePath string, srcDir, stepOutDir string) (*Report, error) {
	rep := &Report{
		Pass: true,
	}

	codeData, err := ReadOSSFile(codePath)
	if err != nil {
		rep.Pass = false
		rep.Message = fmt.Sprintf("failed to get code file, path:%s, err: %s", codePath, err)

		return rep, nil
	}

	id, err := idDispatcher.Get()
	if err != nil {
		// Too many verification items are running at the same time,
		// an error is returned, and the upper layer will retry
		return nil, fmt.Errorf("too many validations running at the same time: %w", err)
	}
	defer idDispatcher.Release(id)

	files, err := custom.Action.GetFiles(srcDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get action files, err: %w", err)
	}

	pl := &pipeline.Pipeline{
		Steps: []pipeline.Step{
			*custom.Action.ToStep(),
		},
		Files: append(files, pipeline.File{
			Name:    "code",
			Content: codeData,
		}),
	}

	res, msg, pass, err := execute(id, pl, stepOutDir)
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
	rep.Pass = pass
	rep.Message = msg

	return rep, nil
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

func execute(id int, pl *pipeline.Pipeline, stepOutDir string) (
	res *pipeline.Result,
	message string,
	pass bool,
	err error,
) {
	executor, err := pipeline.NewExecutor(id)
	if err != nil {
		return
	}
	defer func(executor *pipeline.Executor) {
		if e := executor.Clean(); e != nil {
			err = e
		}
	}(executor)
	res, err = executor.Exec(*pl)
	if err != nil {
		return
	}

	if err = StepOutToOSS(executor.StepOutDir(), stepOutDir); err != nil {
		return
	}
	if data, e := executor.ReadFile("pass"); e == nil && strings.TrimSpace(string(data)) == "true" {
		pass = true
	} else {
		pass = false
	}

	file, _ := executor.ReadFile("message")
	message = string(file)

	return
}

func StepOutToOSS(localDir, ossDir string) error {
	files, err := os.ReadDir(localDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if err = ossClient.PutLocalTextFile(context.Background(), path.Join(localDir, name), path.Join(ossDir, name)); err != nil {
			return err
		}
	}

	return nil
}
