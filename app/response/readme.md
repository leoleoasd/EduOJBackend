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

#### AdminCreateUser

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|     CONFLICT_EMAIL     |        邮箱重复        |
|    CONFLICT_USERNAME   |       用户名重复       |

#### AdminUpdateUser

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到指定user   |
|     CONFLICT_EMAIL     |        邮箱重复        |
|    CONFLICT_USERNAME   |       用户名重复       |

#### AdminDeleteUser

|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到指定user   |

#### AdminGetUser

|         message        |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到指定user   |

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

### CreateProblem

### GetProblem
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到指定problem  |

### GetRandomProblem
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到公开problem  |


### GetProblemAttachmentFile
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|         NOT_FOUND       |     错误的problem     |
|   ATTACHMENT_NOT_FOUND  |     无法找到指定附件    |

### GetProblems
|         message         |         结果         |
|:-----------------------:|:-------------------:|
|     INVALID_STATUS      |     无效的状态设置     |

### UpdateProblem
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到指定problem  |

### DeleteProblem
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到指定problem  |

### CreateTestCase
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     错误的problem     |
|      INVALID_FILE       |        缺少文件       |

### UpdateTestCase
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    无法找到problem    |
|   TEST_CASE_NOT_FOUND   |   无法找到test case   |


### DeleteTestCase
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    无法找到problem    |
|   TEST_CASE_NOT_FOUND   |   无法找到test case   |

### DeleteTestCases
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    无法找到problem    |

### GetTestCaseInputFile
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    无法找到problem    |
|   TEST_CASE_NOT_FOUND   |   无法找到test case   |

### GetTestCaseOutputFile
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |    无法找到problem    |
|   TEST_CASE_NOT_FOUND   |   无法找到test case   |

## Image
### CreateImage
|     code     |  结果   |
|:------------:|:------:|
| ILLEGAL_TYPE | 类型非法 |

## Submission

### CreateSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     错误的problem     |
|     INVALID_LANGUAGE    |       无效的语言       |
|       INVALID_FILE      |        缺少文件       |

### GetSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到submission   |

### GetSubmissions

### GetSubmissionCode
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到submission   |

### GetRunOutput
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|  SUBMISSION_NOT_FOUND   |   无法找到submission   |
|        NOT_FOUND        |      无法找到run       |
|   JUDGEMENT_UNFINISHED  |       评测未完成       |

### GetRunCompilerOutput
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|  SUBMISSION_NOT_FOUND   |   无法找到submission   |
|        NOT_FOUND        |      无法找到run       |
|   JUDGEMENT_UNFINISHED  |       评测未完成       |

### GetRunComparerOutput
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|  SUBMISSION_NOT_FOUND   |   无法找到submission   |
|        NOT_FOUND        |      无法找到run       |
|   JUDGEMENT_UNFINISHED  |       评测未完成       |

# Judger
## UpdateRun

|      message      |                  结果                  |
|:-----------------:|:-------------------------------------:|
|   WRONG_RUN_ID    | 发起请求的judger与获取道当前run的judger不同 |
| ALREADY_SUBMITTED |          一个run被提交了两次结果          |

## Class

### CreateClass

### GetClass
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |

### GetClassesIManage

### GetClassesITake

### UpdateClass
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |

### RefreshInviteCode
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |

### AddStudents
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |

### DeleteStudents
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |

### JoinClass
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     无法找到class      |
|    WRONG_INVITE_CODE    |      错误的邀请码       |
|    ALREADY_IN_CLASS     |   用户已是该class学生   |

### DeleteClass

## ProblemSet

### CreateProblemSet
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|      CLASS_NOT_FOUND    |     无法找到class      |

### CloneProblemSet
|            message           |          结果           |
|:----------------------------:|:----------------------:|
|       CLASS_NOT_FOUND        |       无法找到class      |
|      SOURCE_NOT_FOUND        | 无法找到复制源problem set |

### GetProblemSet
|            message           |          结果           |
|:----------------------------:|:----------------------:|
|           NOT_FOUND          |    无法找到problem set   |
|        CLASS_NOT_FOUND       |        无法找到class     |

### UpdateProblemSet
|            message           |          结果           |
|:----------------------------:|:----------------------:|
|           NOT_FOUND          |    无法找到problem set   |

### AddProblemsToSet
|            message           |          结果           |
|:----------------------------:|:----------------------:|
|           NOT_FOUND          |    无法找到problem set   |

### DeleteProblemsFromSet
|            message           |          结果           |
|:----------------------------:|:----------------------:|
|           NOT_FOUND          |    无法找到problem set   |

### DeleteProblemSet
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|     CLASS_NOT_FOUND     |     无法找到class      |

### GetProblemSetProblem
|            message           |            结果           |
|:----------------------------:|:-------------------------:|
|           NOT_FOUND          |       无法找到problem      |
|     PROBLEM_SET_NOT_FOUND    | 无法找到problem set或 class |

### GetProblemSetProblemInputFile
|            message           |            结果           |
|:----------------------------:|:-------------------------:|
|           NOT_FOUND          |       无法找到problem      |
|     PROBLEM_SET_NOT_FOUND    | 无法找到problem set或 class |

### GetProblemSetProblemOutputFile
|            message           |            结果           |
|:----------------------------:|:-------------------------:|
|           NOT_FOUND          |       无法找到problem      |
|     PROBLEM_SET_NOT_FOUND    | 无法找到problem set或 class |

### RefreshGrades
|            message           |            结果           |
|:----------------------------:|:-------------------------:|
|           NOT_FOUND          | 无法找到problem set或 class |

## ProblemSetSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|  PROBLEM_SET_NOT_FOUND  |  无法找到problem set  |


### ProblemSetCreateSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |     错误的problem     |
|     INVALID_LANGUAGE    |       无效的语言       |
|       INVALID_FILE      |        缺少文件       |

### ProblemSetGetSubmission
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到submission   |

### ProblemSetGetSubmissions

### ProblemSetGetSubmissionCode
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |   无法找到submission   |

### ProblemSetGetRunOutput
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |      无法找到run       |
|   JUDGEMENT_UNFINISHED  |       评测未完成       |

### ProblemSetGetRunInput

### ProblemSetGetRunCompilerOutput
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |      无法找到run       |
|   JUDGEMENT_UNFINISHED  |       评测未完成       |

### ProblemSetGetRunComparerOutput
|         message         |         结果          |
|:-----------------------:|:--------------------:|
|        NOT_FOUND        |      无法找到run       |
|   JUDGEMENT_UNFINISHED  |       评测未完成       |
