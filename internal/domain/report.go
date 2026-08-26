package domain

import (
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportReason represents the predefined reasons for reporting a post
type ReportReason string

const (
	ReportReasonUnder18              ReportReason = "problem_involving_someone_under_18"
	ReportReasonBullyingHarassment   ReportReason = "bullying_harassment_or_abuse"
	ReportReasonSuicideSelfHarm      ReportReason = "suicide_or_self_harm"
	ReportReasonViolentHateful       ReportReason = "violent_hateful_or_disturbing_content"
	ReportReasonSellingRestricted    ReportReason = "selling_or_promoting_restricted_items"
	ReportReasonAdultContent         ReportReason = "adult_content"
	ReportReasonScamFraud            ReportReason = "scam_fraud_or_false_information"
	ReportReasonIntellectualProperty ReportReason = "intellectual_property"
	ReportReasonDontWantToSee        ReportReason = "i_dont_want_to_see_this"
)

// Report represents a user report on a post
type Report struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	PostID    primitive.ObjectID `bson:"post_id"`
	UserID    primitive.ObjectID `bson:"user_id"` // Reporter (hidden from post creator)
	Reason    ReportReason       `bson:"reason"`
	Comment   *string            `bson:"comment,omitempty"` // Optional comment from reporter
	CreatedAt time.Time          `bson:"created_at"`
}

// Sentinel errors for reporting
var (
	ErrReportAlreadyExists = &AppError{Code: "REPORT_ALREADY_EXISTS", Message: "You have already reported this post", HTTPStatus: http.StatusConflict}
	ErrReportLimitExceeded = &AppError{Code: "REPORT_LIMIT_EXCEEDED", Message: "You have reached the maximum number of reports per day", HTTPStatus: http.StatusTooManyRequests}
	ErrInvalidReportReason = &AppError{Code: "INVALID_REPORT_REASON", Message: "The provided report reason is invalid", HTTPStatus: http.StatusBadRequest}
	ErrCannotReportOwnPost = &AppError{Code: "CANNOT_REPORT_OWN_POST", Message: "You cannot report your own post", HTTPStatus: http.StatusBadRequest}
)

// IsValidReason checks if the provided reason is valid
func IsValidReason(reason string) bool {
	validReasons := []ReportReason{
		ReportReasonUnder18,
		ReportReasonBullyingHarassment,
		ReportReasonSuicideSelfHarm,
		ReportReasonViolentHateful,
		ReportReasonSellingRestricted,
		ReportReasonAdultContent,
		ReportReasonScamFraud,
		ReportReasonIntellectualProperty,
		ReportReasonDontWantToSee,
	}

	for _, r := range validReasons {
		if string(r) == reason {
			return true
		}
	}
	return false
}
