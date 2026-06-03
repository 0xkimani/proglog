package auth

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
"github.com/casbin/casbin/v2"
)

// Wraps casbin's Enforcer function
type Authorizer struct {
	enforcer *casbin.Enforcer
}

// New function's model and policy arguments are paths
// to the files where you've defined the model - which will
// configure Casbin's authorization mechanism {ACL} and policy {CSV file
// containing ACL table}
func New(model, policy string) *Authorizer {
    enforcer, err := casbin.NewEnforcer(model, policy)
    if err != nil {
        panic(err) // or return the error if you prefer
    }
    return &Authorizer{
        enforcer: enforcer,
    }
}

// Authorize defers to Casbin's Enforcer function. This function
// returns whether the given subject is permitted to run the given action
// on the given object based on the model and policy you configure Casbin with
func (a *Authorizer) Authorize(subject, object, action string) error {
    ok, err := a.enforcer.Enforce(subject, object, action)
    if err != nil {
        return err
    }
    if !ok {
        msg := fmt.Sprintf(
            "%s not permitted to %s to %s",
            subject,
            action,
            object,
        )
        st := status.New(codes.PermissionDenied, msg)
        return st.Err()
    }
    return nil
}
