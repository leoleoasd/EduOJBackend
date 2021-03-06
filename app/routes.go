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

	auth := api.Group("/auth", middleware.Auth)
	auth.POST("/login", controller.Login).Name = "auth.login"
	auth.GET("/login/webauthn", controller.BeginLogin).Name = "auth.webauthn.beginLogin"
	auth.POST("/login/webauthn", controller.FinishLogin).Name = "auth.webauthn.finishLogin"
	auth.POST("/register", controller.Register).Name = "auth.register"
	auth.GET("/email_registered", controller.EmailRegistered).Name = "auth.emailRegistered"

	admin := api.Group("/admin", middleware.Logged)
	admin.POST("/user",
		controller.AdminCreateUser, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "manage_user"})).Name = "admin.user.createUser"
	admin.PUT("/user/:id",
		controller.AdminUpdateUser, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "manage_user"})).Name = "admin.user.updateUser"
	admin.DELETE("/user/:id",
		controller.AdminDeleteUser, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "manage_user"})).Name = "admin.user.deleteUser"
	admin.GET("/user/:id",
		controller.AdminGetUser, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "read_user"})).Name = "admin.user.getUser"
	admin.GET("/users",
		controller.AdminGetUsers, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "read_user"})).Name = "admin.user.getUsers"

	api.GET("/user/me", controller.GetMe, middleware.Logged).Name = "user.getMe"
	api.PUT("/user/me", controller.UpdateMe, middleware.Logged).Name = "user.updateMe"
	api.GET("/user/:id", controller.GetUser).Name = "user.getUser"
	api.GET("/users", controller.GetUsers).Name = "user.getUsers"

	api.GET("/webauthn/register", controller.BeginRegistration, middleware.Logged).Name = "webauthn.BeginRegister"
	api.POST("/webauthn/register", controller.FinishRegistration, middleware.Logged).Name = "webauthn.FinishRegister"

	api.POST("/user/change_password", controller.ChangePassword, middleware.Logged).Name = "user.changePassword"

	api.GET("/image/:id", controller.GetImage).Name = "image.getImage"
	api.POST("/image", controller.CreateImage, middleware.Logged).Name = "image.createImage"

	admin.POST("/problem",
		controller.CreateProblem, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "create_problem"})).Name = "problem.createProblem"
	admin.PUT("/problem/:id",
		controller.UpdateProblem, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "update_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "update_problem"},
		})).Name = "problem.updateProblem"
	admin.DELETE("/problem/:id",
		controller.DeleteProblem, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "delete_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "delete_problem"},
		})).Name = "problem.deleteProblem"

	api.GET("/problem/random", controller.GetRandomProblem, middleware.AllowGuest).Name = "problem.getRandomProblem"
	api.GET("/problem/:id", controller.GetProblem, middleware.AllowGuest).Name = "problem.getProblem"
	api.GET("/problems", controller.GetProblems, middleware.AllowGuest).Name = "problem.getProblems"

	api.GET("/problem/:id/attachment_file", controller.GetProblemAttachmentFile, middleware.AllowGuest).Name = "problem.getProblemAttachmentFile"

	admin.POST("/problem/:id/test_case",
		controller.CreateTestCase,
		middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "update_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "update_problem"},
		})).Name = "problem.createTestCase"
	admin.PUT("/problem/:id/test_case/:test_case_id",
		controller.UpdateTestCase,
		middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "update_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "update_problem"},
		})).Name = "problem.updateTestCase"
	admin.DELETE("/problem/:id/test_case/all",
		controller.DeleteTestCases,
		middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "update_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "update_problem"},
		})).Name = "problem.deleteTestCases"
	admin.DELETE("/problem/:id/test_case/:test_case_id",
		controller.DeleteTestCase,
		middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "update_problem", T: "problem"},
			B: middleware.UnscopedPermission{P: "update_problem"},
		})).Name = "problem.deleteTestCase"

	api.GET("/problem/:id/test_case/:test_case_id/input_file",
		controller.GetTestCaseInputFile,
		middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.CustomPermission{
				F: middleware.IsTestCaseSample,
			},
			B: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "read_problem_secret", T: "problem"},
				B: middleware.UnscopedPermission{P: "read_problem_secret"},
			},
		})).Name = "problem.getTestCaseInputFile"
	api.GET("/problem/:id/test_case/:test_case_id/output_file",
		controller.GetTestCaseOutputFile,
		middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.CustomPermission{
				F: middleware.IsTestCaseSample,
			},
			B: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "read_problem_secret", T: "problem"},
				B: middleware.UnscopedPermission{P: "read_problem_secret"},
			},
		})).Name = "problem.getTestCaseOutputFile"

	api.GET("/user/:id/problem_info", controller.GetUserProblemInfo).Name = "problem.getUserProblemInfo"

	api.POST("/problem/:pid/submission", controller.CreateSubmission, middleware.Logged).Name = "submission.createSubmission"
	api.GET("/submission/:id", controller.GetSubmission, middleware.Logged).Name = "submission.getSubmission"
	api.GET("/submissions", controller.GetSubmissions, middleware.Logged).Name = "submission.getSubmissions"

	api.GET("/submission/:id/code", controller.GetSubmissionCode, middleware.Logged).Name = "submission.getSubmissionCode"
	api.GET("/submission/:id/run/:run_id/output", controller.GetRunOutput, middleware.Logged).Name = "submission.getRunOutput"
	api.GET("/submission/:id/run/:run_id/input", controller.GetRunInput, middleware.Logged).Name = "submission.getRunInput"
	api.GET("/submission/:id/run/:run_id/compiler_output", controller.GetRunCompilerOutput, middleware.Logged).Name = "submission.getRunCompilerOutput"
	api.GET("/submission/:id/run/:run_id/comparer_output", controller.GetRunComparerOutput, middleware.Logged).Name = "submission.getRunComparerOutput"

	api.POST("/problem_set/:problem_set_id/problem/:pid/submission",
		controller.ProblemSetCreateSubmission, middleware.Logged).Name = "problemSet.createSubmission"
	api.GET("/problem_set/:problem_set_id/submission/:id",
		controller.ProblemSetGetSubmission, middleware.Logged).Name = "problemSet.getSubmission"
	api.GET("/problem_set/:problem_set_id/submissions",
		controller.ProblemSetGetSubmissions, middleware.Logged).Name = "problemSet.getSubmissions"

	api.GET("/problem_set/:problem_set_id/submission/:id/code",
		controller.ProblemSetGetSubmissionCode, middleware.Logged).Name = "problemSet.getSubmissionCode"
	api.GET("/problem_set/:problem_set_id/submission/:submission_id/run/:id/output",
		controller.ProblemSetGetRunOutput, middleware.Logged).Name = "problemSet.getRunOutput"
	api.GET("/problem_set/:problem_set_id/submission/:submission_id/run/:id/input",
		controller.ProblemSetGetRunInput, middleware.Logged).Name = "problemSet.getRunInput"
	api.GET("/problem_set/:problem_set_id/submission/:submission_id/run/:id/compiler_output",
		controller.ProblemSetGetRunCompilerOutput, middleware.Logged).Name = "problemSet.getRunCompilerOutput"
	api.GET("/problem_set/:problem_set_id/submission/:submission_id/run/:id/comparer_output",
		controller.ProblemSetGetRunComparerOutput, middleware.Logged).Name = "problemSet.getRunComparerOutput"

	admin.GET("/logs",
		controller.AdminGetLogs, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "read_logs"})).Name = "admin.getLogs"

	judger := e.Group("/judger", middleware.Judger)

	judger.GET("/script/:name", controller.JudgerGetScript).Name = "judger.getScript"
	judger.GET("/task", controller.GetTask).Name = "judger.getTask"
	judger.PUT("/run/:id", controller.UpdateRun).Name = "judger.updateRun"

	api.POST("/class",
		controller.CreateClass, middleware.Logged, middleware.HasPermission(middleware.UnscopedPermission{P: "manage_class"})).Name = "class.createClass"
	api.GET("/class/:id",
		controller.GetClass, middleware.Logged).Name = "class.getClass"
	api.GET("/user/me/managing_classes",
		controller.GetClassesIManage, middleware.Logged).Name = "class.getClassesIManage"
	api.GET("/user/me/taking_classes",
		controller.GetClassesITake, middleware.Logged).Name = "class.getClassesITake"
	api.PUT("/class/:id",
		controller.UpdateClass, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_class", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_class"},
		})).Name = "class.updateClass"
	api.PUT("/class/:id/invite_code",
		controller.RefreshInviteCode, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_class", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_class"},
		})).Name = "class.refreshInviteCode"
	api.POST("/class/:id/students",
		controller.AddStudents, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_class", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_class"},
		})).Name = "class.addStudents"
	api.DELETE("/class/:id/students",
		controller.DeleteStudents, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_class", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_class"},
		})).Name = "class.deleteStudents"
	api.POST("/class/:id/join", controller.JoinClass, middleware.Logged).Name = "class.joinClass"
	api.DELETE("/class/:id",
		controller.DeleteClass, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_class", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_class"},
		})).Name = "class.deleteClass"

	api.POST("/class/:id/problem_set",
		controller.CreateProblemSet, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		})).Name = "problemSet.createProblemSet"
	api.POST("/class/:id/problem_set/clone",
		controller.CloneProblemSet, middleware.Logged, middleware.HasPermission(middleware.AndPermission{
			A: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class"},
				B: middleware.UnscopedPermission{P: "manage_problem_sets"},
			},
			B: middleware.OrPermission{
				A: middleware.ScopedPermission{P: "clone_problem_sets", T: "class"},
				B: middleware.UnscopedPermission{P: "clone_problem_sets"},
			},
		})).Name = "problemSet.cloneProblemSet"
	api.GET("/class/:class_id/problem_set/:id",
		controller.GetProblemSet, middleware.Logged).Name = "problemSet.getProblemSet"
	api.PUT("/class/:class_id/problem_set/:id",
		controller.UpdateProblemSet, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		})).Name = "problemSet.updateProblemSet"
	api.POST("/class/:class_id/problem_set/:id/problems",
		controller.AddProblemsToSet, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		})).Name = "problemSet.addProblemsToSet"
	api.DELETE("/class/:class_id/problem_set/:id/problems",
		controller.DeleteProblemsFromSet, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		})).Name = "problemSet.deleteProblemsFromSet"
	api.DELETE("/class/:class_id/problem_set/:id",
		controller.DeleteProblemSet, middleware.Logged, middleware.HasPermission(middleware.OrPermission{
			A: middleware.ScopedPermission{P: "manage_problem_sets", T: "class", IdFieldName: "class_id"},
			B: middleware.UnscopedPermission{P: "manage_problem_sets"},
		})).Name = "problemSet.deleteProblemSet"

	script := admin.Group("/script", middleware.HasPermission(middleware.UnscopedPermission{P: "manage_scripts"}))
	script.POST("", controller.CreateScript).Name = "script.createScript"
	script.GET("/:name", controller.GetScript).Name = "script.getScript"
	script.GET("/:name/file", controller.GetScriptFile).Name = "script.getScriptFile"
	script.GET("s", controller.GetScripts).Name = "script.getScripts"
	script.PUT("/:name", controller.UpdateScript).Name = "script.updateScript"
	script.DELETE("/:name", controller.DeleteScript).Name = "script.deleteScript"

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
