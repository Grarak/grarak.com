package mandy

type MandyErrorCode int

const CODE_NO_ERROR MandyErrorCode = 0
const CODE_UNKNOWN_ERROR MandyErrorCode = 1
const CODE_API_INVALID MandyErrorCode = 2
const CODE_USERNAME_TAKEN MandyErrorCode = 3
const CODE_PASSWORD_SHORT MandyErrorCode = 4
const CODE_USERNAME_SHORT MandyErrorCode = 5
const cODE_USERNAME_INVALID MandyErrorCode = 6
const CODE_USERNAME_PASSWORD_INVALID MandyErrorCode = 7
const CODE_MERGE_FAILED MandyErrorCode = 8
const CODE_SET_CONFLICTION_FAILED MandyErrorCode = 9
const CODE_REVERT_FAILED MandyErrorCode = 10
const CODE_SUBMIT_FAILED MandyErrorCode = 11
const CODE_GET_USERS_FAILED MandyErrorCode = 12
const CODE_SET_VERIFICATION_FAILED MandyErrorCode = 13
const CODE_SET_MODERATION_FAILED MandyErrorCode = 13
const CODE_REMOVE_USER_FAILED MandyErrorCode = 14
const CODE_GET_NOTIFICATION_ACTIVITIES_FAILED MandyErrorCode = 15
