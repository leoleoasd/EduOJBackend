package app

import (
	"github.com/EduOJ/backend/app/controller"
	"github.com/EduOJ/backend/app/middleware"
	"github.com/EduOJ/backend/base/log"
	"github.com/EduOJ/backend/base/utils"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"net/http"
	"net/http/pprof"
)

func Register(e *echo.Echo) {
	utils.InitOrigin()
	e.Use(middleware.Recover)
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     utils.Origins,
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
	}))
	api := e.Group("/api", middleware.Authentication)

	// judger APIs
	judger := e.Group("", middleware.Judger)
	judger.PUT("/judger/run/:id", controller.UpdateRun,
		middleware.ValidateParams(map[string]string{
			"id": "NOT_FOUND",
		}),
	).Name = "judger.updateRun"
	judger.GET("/judger/script/:name", controller.GetScript).Name = "judger.getScript"
	judger.GET("/judger/task", controller.GetTask).Name = "judger.getTask"

	// auth APIs
	auth := api.Group("", middleware.Auth)
	auth.POST("/auth/login", controller.Login).Name = "auth.login"
	auth.GET("/auth/login/webauthn", controller.BeginLogin).Name = "auth.webauthn.beginLogin"
	auth.POST("/auth/login/webauthn", controller.FinishLogin).Name = "auth.webauthn.finishLogin"
	auth.POST("/auth/register", controller.Register).Name = "auth.register"
	auth.GET("/auth/email_registered", controller.EmailRegistered).Name = "auth.emailRegistered"

	// user APIs
	user := api.Group("", middleware.Logged)
	readUser := api.Group("",
		middleware.Logged,
		middleware.HasPermission(middleware.UnscopedPermission{P: "read_user"}),
	)
	manageUsers := api.Group("",
		middleware.Logged,
		middleware.HasPermission(middleware.UnscopedPermission{P: "manage_user"}),
	)
	user.GET("/user/me", controller.GetMe).Name = "user.getMe"
	user.PUT("/user/me", controller.UpdateMe).Name = "user.updateMe"
	api.GET("/user/:id", controller.GetUser).Name = "user.getUser"
	user.GET("/user/me/managing_classes", controller.GetClassesIManage).Name = "user.getClassesIManage"
	user.GET("/user/me/taking_classes", controller.GetClassesITake).Name = "user.getClassesITake"
	user.GET("/user/:id/problem_info", controller.GetUserProblemInfo).Name = "user.getUserProblemInfo"
	user.GET("/users", controller.GetUsers).Name = "user.getUsers"
	user.POST("/user/change_password", controller.ChangePassword).Name = "user.changePassword"
	readUser.GET("/admin/user/:id", controller.AdminGetUser).Name = "admin.user.getUser"
	readUser.GET("/admin/users", controller.AdminGetUsers).Name = "admin.user.getUsers"
	manageUsers.POST("/admin/user", controller.AdminCreateUser).Name = "admin.user.createUser"
	manageUsers.PUT("/admin/user/:id", controller.AdminUpdateUser).Name = "admin.user.updateUser"
	manageUsers.DELETE("/admin/user/:id", controller.AdminDeleteUser).Name = "admin.user.deleteUser"

	// webauthn APIs
	webauthn := api.Group("",
		middleware.Logged,
	)
	webauthn.GET("/webauthn/register", controller.BeginRegistration).Name = "webauthn.BeginRegister"
	webauthn.POST("/webauthn/register", controller.FinishRegistration).Name = "webauthn.FinishRegister"

	// image APIs
	image := api.Group("")
	image.GET("/image/:id", controller.GetImage).Name = "image.getImage"
	image.POST("/image", controller.CreateImage, middleware.Logged).Name = "image.createImage"

	// problem APIs
	problem := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id": "NOT_FOUND",
		}),
		middleware.AllowGuest)
	readProblemSecret := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":           "NOT_FOUND",
			"test_case_id": "TEST_CASE_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.CustomPermission{
				F: middleware.IsTestCaseSample,
			},
			B: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "read_problem_secrets", T: "problem"},
				B: middleware.UnscopedPermission{P: "read_problem_secrets"},
			},
		}),
	)
	updateProblem := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":           "NOT_FOUND",
			"test_case_id": "TEST_CASE_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "update_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "update_problem"},
		}),
	)
	api.POST("/admin/problem", controller.CreateProblem,
		middleware.Logged,
		middleware.HasPermission(middleware.UnscopedPermission{P: "create_problem"}),
	).Name = "problem.createProblem"
	api.DELETE("/problem/:id", controller.DeleteProblem,
		middleware.ValidateParams(map[string]string{
			"id": "NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "delete_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "delete_problem"},
		}),
	).Name = "problem.deleteProblem"
	problem.GET("/problem/random", controller.GetRandomProblem).Name = "problem.getRandomProblem"
	problem.GET("/problem/:id", controller.GetProblem).Name = "problem.getProblem"
	problem.GET("/problems", controller.GetProblems).Name = "problem.getProblems"
	problem.GET("/problem/:id/attachment_file", controller.GetProblemAttachmentFile).Name = "problem.getProblemAttachmentFile"
	readProblemSecret.GET("/problem/:id/test_case/:test_case_id/input_file", controller.GetTestCaseInputFile).Name = "problem.getTestCaseInputFile"
	readProblemSecret.GET("/problem/:id/test_case/:test_case_id/output_file", controller.GetTestCaseOutputFile).Name = "problem.getTestCaseOutputFile"
	updateProblem.PUT("/admin/problem/:id", controller.UpdateProblem).Name = "problem.updateProblem"
	updateProblem.POST("/admin/problem/:id/test_case", controller.CreateTestCase).Name = "problem.createTestCase"
	updateProblem.PUT("/admin/problem/:id/test_case/:test_case_id", controller.UpdateTestCase).Name = "problem.updateTestCase"
	updateProblem.DELETE("/admin/problem/:id/test_case/all", controller.DeleteTestCases).Name = "problem.deleteTestCases"
	updateProblem.DELETE("/admin/problem/:id/test_case/:test_case_id", controller.DeleteTestCase).Name = "problem.deleteTestCase"

	// submission APIs
	submission := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":            "NOT_FOUND",
			"submission_id": "SUBMISSION_NOT_FOUND",
			"problem_id":    "PROBLEM_NOT_FOUND",
		}),
		middleware.Logged,
	)
	submission.POST("/problem/:problem_id/submission", controller.CreateSubmission).Name = "submission.createSubmission"
	submission.GET("/submission/:id", controller.GetSubmission).Name = "submission.getSubmission"
	submission.GET("/submissions", controller.GetSubmissions, middleware.Logged).Name = "submission.getSubmissions"
	submission.GET("/submission/:id/code", controller.GetSubmissionCode, middleware.Logged).Name = "submission.getSubmissionCode"
	submission.GET("/submission/:submission_id/run/:id/output", controller.GetRunOutput, middleware.Logged).Name = "submission.getRunOutput"
	submission.GET("/submission/:submission_id/run/:id/input", controller.GetRunInput, middleware.Logged).Name = "submission.getRunInput"
	submission.GET("/submission/:submission_id/run/:id/compiler_output", controller.GetRunCompilerOutput, middleware.Logged).Name = "submission.getRunCompilerOutput"
	submission.GET("/submission/:submission_id/run/:id/comparer_output", controller.GetRunComparerOutput, middleware.Logged).Name = "submission.getRunComparerOutput"

	// log API
	api.GET("/admin/logs", controller.AdminGetLogs,
		middleware.Logged,
		middleware.HasPermission(middleware.UnscopedPermission{P: "read_logs"}),
	).Name = "admin.getLogs"

	// class APIs
	class := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id": "NOT_FOUND",
		}),
		middleware.Logged,
	)
	manageClass := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id": "NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_class", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_class"},
		}),
	)
	api.POST("/class", controller.CreateClass,
		middleware.Logged,
		middleware.HasPermission(middleware.UnscopedPermission{P: "manage_class"}),
	).Name = "class.createClass"
	class.GET("/class/:id", controller.GetClass, middleware.Logged).Name = "class.getClass"
	class.POST("/class/:id/join", controller.JoinClass, middleware.Logged).Name = "class.joinClass"
	manageClass.PUT("/class/:id", controller.UpdateClass).Name = "class.updateClass"
	manageClass.PUT("/class/:id/invite_code", controller.RefreshInviteCode).Name = "class.refreshInviteCode"
	manageClass.POST("/class/:id/students", controller.AddStudents).Name = "class.addStudents"
	manageClass.DELETE("/class/:id/students", controller.DeleteStudents).Name = "class.deleteStudents"
	manageClass.DELETE("/class/:id", controller.DeleteClass).Name = "class.deleteClass"

	// problem set APIs
	createProblemSet := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id": "NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		}),
	)
	manageProblemSet := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":       "NOT_FOUND",
			"class_id": "CLASS_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		}),
	)
	problemSetProblem := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":             "NOT_FOUND",
			"class_id":       "CLASS_NOT_FOUND",
			"problem_set_id": "PROBLEM_SET_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
				B: middleware.UnscopedPermission{P: "manage_problem_sets"},
			},
			B: middleware.CustomPermission{F: middleware.ProblemSetStarted},
		}),
	)
	problemSetGrades := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":       "NOT_FOUND",
			"class_id": "CLASS_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "read_answers", T: "class", IdFieldName: "class_id"},
			B: middleware.UnscopedPermission{P: "read_answers"},
		}),
	)
	api.GET("/class/:class_id/problem_set/:problem_set_id", controller.GetProblemSet,
		middleware.ValidateParams(map[string]string{
			"id":       "NOT_FOUND",
			"class_id": "CLASS_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
				B: middleware.UnscopedPermission{P: "manage_problem_sets"},
			},
			B: middleware.CustomPermission{F: middleware.ProblemSetStarted},
		}),
	).Name = "problemSet.getProblemSet"
	createProblemSet.POST("/class/:id/problem_set", controller.CreateProblemSet).Name = "problemSet.createProblemSet"
	createProblemSet.POST("/class/:id/problem_set/clone", controller.CloneProblemSet).Name = "problemSet.cloneProblemSet" // TODO: add clone_problem_sets perm check
	manageProblemSet.PUT("/class/:class_id/problem_set/:problem_set_id", controller.UpdateProblemSet).Name = "problemSet.updateProblemSet"
	manageProblemSet.POST("/class/:class_id/problem_set/:id/problems", controller.AddProblemsToSet).Name = "problemSet.addProblemsToSet"
	manageProblemSet.DELETE("/class/:class_id/problem_set/:id/problems", controller.DeleteProblemsFromSet).Name = "problemSet.deleteProblemsFromSet"
	manageProblemSet.DELETE("/class/:class_id/problem_set/:problem_set_id", controller.DeleteProblemSet).Name = "problemSet.deleteProblemSet"
	problemSetProblem.GET("/class/:class_id/problem_set/:problem_set_id/problem/:id", controller.GetProblemSetProblem).Name = "problemSet.getProblemSetProblem"
	problemSetProblem.GET("/class/:class_id/problem_set/:problem_set_id/problem/:id/test_case/:test_case_id/input_file", controller.GetProblemSetProblemInputFile,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.CustomPermission{
				F: middleware.IsTestCaseSampleProblemSet,
			},
			B: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "read_problem_secrets", T: "class", IdFieldName: "class_id"},
				B: middleware.UnscopedPermission{P: "read_problem_secrets"},
			},
		})).Name = "problemSet.getProblemSetProblemInputFile" // TODO: refactor these middleware into groups
	problemSetProblem.GET("/class/:class_id/problem_set/:problem_set_id/problem/:id/test_case/:test_case_id/output_file", controller.GetProblemSetProblemOutputFile,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.CustomPermission{
				F: middleware.IsTestCaseSampleProblemSet,
			},
			B: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "read_problem_secrets", T: "class", IdFieldName: "class_id"},
				B: middleware.UnscopedPermission{P: "read_problem_secrets"},
			},
		})).Name = "problemSet.getProblemSetProblemOutputFile"
	problemSetGrades.GET("/class/:class_id/problem_set/:id/grades", controller.Todo).Name = "problemSet.GetGrades"
	problemSetGrades.POST("/class/:class_id/problem_set/:id/grades/refresh", controller.RefreshGrades).Name = "problemSet.RefreshGrades"

	// problem set submission APIs
	problemSetSubmission := api.Group("",
		middleware.ValidateParams(map[string]string{
			"id":             "NOT_FOUND",
			"class_id":       "CLASS_NOT_FOUND",
			"problem_set_id": "PROBLEM_SET_NOT_FOUND",
			"submission_id":  "SUBMISSION_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.OrPermission{
			A: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "read_answers", T: "class", IdFieldName: "class_id"},
				B: middleware.UnscopedPermission{P: "read_answers"},
			},
			B: middleware.CustomPermission{F: middleware.ProblemSetStarted},
		}),
	)
	api.POST("/class/:class_id/problem_set/:problem_set_id/problem/:problem_id/submission", controller.ProblemSetCreateSubmission,
		middleware.ValidateParams(map[string]string{
			"class_id":       "CLASS_NOT_FOUND",
			"problem_set_id": "PROBLEM_SET_NOT_FOUND",
			"problem_id":     "PROBLEM_NOT_FOUND",
		}),
		middleware.Logged,
		middleware.HasPermission(middleware.CustomPermission{F: middleware.ProblemSetStarted}),
	).Name = "problemSet.createSubmission"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submission/:id", controller.ProblemSetGetSubmission).Name = "problemSet.getSubmission"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submissions", controller.ProblemSetGetSubmissions).Name = "problemSet.getSubmissions"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submission/:id/code", controller.ProblemSetGetSubmissionCode).Name = "problemSet.getSubmissionCode"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submission/:submission_id/run/:id/output", controller.ProblemSetGetRunOutput).Name = "problemSet.getRunOutput"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submission/:submission_id/run/:id/input", controller.ProblemSetGetRunInput).Name = "problemSet.getRunInput"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submission/:submission_id/run/:id/compiler_output", controller.ProblemSetGetRunCompilerOutput).Name = "problemSet.getRunCompilerOutput"
	problemSetSubmission.GET("/class/:class_id/problem_set/:problem_set_id/submission/:submission_id/run/:id/comparer_output", controller.ProblemSetGetRunComparerOutput).Name = "problemSet.getRunComparerOutput"

	// pprof APIs
	if viper.GetBool("debug") {
		log.Debugf("Adding pprof handlers. SHOULD NOT BE USED IN PRODUCTION")
		e.Any("/debug/pprof/", func(c echo.Context) error {
			pprof.Index(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/*", func(c echo.Context) error {
			pprof.Index(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/cmdline", func(c echo.Context) error {
			pprof.Cmdline(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/profile", func(c echo.Context) error {
			pprof.Profile(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/symbol", func(c echo.Context) error {
			pprof.Symbol(c.Response().Writer, c.Request())
			return nil
		})
		e.Any("/debug/pprof/trace", func(c echo.Context) error {
			pprof.Trace(c.Response().Writer, c.Request())
			return nil
		})
	}
}
