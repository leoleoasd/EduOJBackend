# API
Common return codes.

|        message         |         结果          |
|:----------------------:|:--------------------:|
|        SUCCESS         |         正常          |
|     INTERNAL_ERROR     |     服务器内部错误      |
|    VALIDATION_ERROR    | 验证错误, 详见error字段 |


# Authentication

|       message        |       结果       |
|:--------------------:|:---------------:|
| AUTH_SESSION_EXPIRED |   session超时    |
|   AUTH_NEED_TOKEN    |   未提供token    |

# Permission

|          message           |         结果          |
|:--------------------------:|:--------------------:|
|     PERMISSION_DENIED      |        没有权限        |

## Auth

### Login

|       message      |         结果          |
|:------------------:|:--------------------:|
|   WRONG_USERNAME   |    错误的用户名/邮箱    |
|   WRONG_PASSWORD   |       密码错误        |

### Register

|          message           |         结果          |
|:--------------------------:|:--------------------:|
|      CONFLICT_EMAIL       |       邮箱重复        |
|     CONFLICT_USERNAME     |       用户名重复       |

### EmailRegistered

|          message           |         结果          |
|:--------------------------:|:--------------------:|
|      EMAIL_REGISTERED      |       邮箱已注册       |

## Admin

### Log

#### AdminGetLogs
|   message     |    结果     |
|:-------------:|:----------:|
| INVALID_LEVEL | 非法的level |

### User
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到指定user   |

#### AdminCreateUser

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|     CONFLICT_EMAIL     |        邮箱重复        |
|    CONFLICT_USERNAME   |       用户名重复       |

#### AdminUpdateUser

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|     CONFLICT_EMAIL     |        邮箱重复        |
|    CONFLICT_USERNAME   |       用户名重复       |

#### AdminDeleteUser

#### AdminGetUser

#### AdminGetUsers
|         message         |         结果         |
|:-----------------------:|:-------------------:|
|      INVALID_ORDER      |     无效的排序设置     |

## User

### GetMe

### UpdateMe

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|     CONFLICT_EMAIL     |        邮箱重复        |
|    CONFLICT_USERNAME   |       用户名重复       |

### GetUser

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到指定user   |

### GetUsers

|         message         |         结果         |
|:-----------------------:|:-------------------:|
|      INVALID_ORDER      |     无效的排序设置     |

### ChangePassword

|       message        |         结果          |
|:--------------------:|:--------------------:|
|    WRONG_PASSWORD    |       密码错误        |

## Problem
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    无法找到problem    |
|   TEST_CASE_NOT_FOUND   |   无法找到test case   |

### CreateProblem

### GetProblem

### GetRandomProblem

### GetProblemAttachmentFile
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|   ATTACHMENT_NOT_FOUND  |     无法找到指定附件    |

### GetProblems
|         message         |         结果         |
|:-----------------------:|:-------------------:|
|     INVALID_STATUS      |     无效的状态设置     |

### UpdateProblem

### DeleteProblem

### CreateTestCase
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|      INVALID_FILE       |        缺少文件       |

### UpdateTestCase

### DeleteTestCase

### DeleteTestCases

### GetTestCaseInputFile

### GetTestCaseOutputFile

## Image
### CreateImage
|     code     |  结果   |
|:------------:|:------:|
| ILLEGAL_TYPE | 类型非法 |

## Submission
|         message         |            结果            |
|:-----------------------:|:-------------------------:|
|  SUBMISSION_NOT_FOUND   |      无法找到submission    |
|        NOT_FOUND        |   无法找到submission或run   |

### CreateSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     错误的problem     |
|     INVALID_LANGUAGE    |       无效的语言       |
|       INVALID_FILE      |        缺少文件       |

### GetSubmission

### GetSubmissions

### GetSubmissionCode

### GetRunOutput

### GetRunCompilerOutput

### GetRunComparerOutput

# Judger

## JudgerGetScript

## UpdateRun

|      message      |                  结果                  |
|:-----------------:|:-------------------------------------:|
|   WRONG_RUN_ID    | 发起请求的judger与获取道当前run的judger不同 |
| ALREADY_SUBMITTED |          一个run被提交了两次结果          |

## Class
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |

### CreateClass

### GetClass

### GetClassesIManage

### GetClassesITake

### UpdateClass

### RefreshInviteCode

### AddStudents

### DeleteStudents

### JoinClass
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|    ALREADY_IN_CLASS     |   用户已是该class学生   |

### DeleteClass

## ProblemSet

|            message           |          结果           |
|:----------------------------:|:----------------------:|
|         CLASS_NOT_FOUND      |       无法找到class      |
|           NOT_FOUND          |    无法找到problem set   |

### CreateProblemSet

### CloneProblemSet
|            message           |          结果           |
|:----------------------------:|:----------------------:|
|      SOURCE_NOT_FOUND        | 无法找到复制源problem set |

### GetProblemSet

### UpdateProblemSet

### AddProblemsToSet

### DeleteProblemsFromSet

### DeleteProblemSet

## ProblemSetSubmission
|         message         |         结果           |
|:-----------------------:|:---------------------:|
|        NOT_FOUND        | 无法找到submission或run |

### ProblemSetCreateSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|  PROBLEM_SET_NOT_FOUND  |  无法找到problem set  |
|        NOT_FOUND        |     错误的problem     |
|     INVALID_LANGUAGE    |       无效的语言       |
|       INVALID_FILE      |        缺少文件       |

### ProblemSetGetSubmission

### ProblemSetGetSubmissions

### ProblemSetGetSubmissionCode

### ProblemSetGetRunOutput

### ProblemSetGetRunCompilerOutput

### ProblemSetGetRunComparerOutput

## Script
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    找不到指定script    |

### CreateScript
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|       INVALID_FILE      |        缺少文件       |

### GetScript

### GetScriptFile

### GetScripts

### UpdateScript

### DeleteScript
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|      SCRIPT_IN_USE      |      脚本仍在使用      |