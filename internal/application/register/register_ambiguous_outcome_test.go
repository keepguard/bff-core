package register

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"syscall"
	"testing"

	authDto "github.com/keepguard/bff-core/internal/adapters/out/http/dto/auth"
	appdto "github.com/keepguard/bff-core/internal/application/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeTimeoutErr struct{}

func (fakeTimeoutErr) Error() string   { return "i/o timeout" }
func (fakeTimeoutErr) Timeout() bool   { return true }
func (fakeTimeoutErr) Temporary() bool { return true }

func wrapURL(err error) error {
	return &url.Error{Op: "Post", URL: "http://ms-auth/api/v1/users/create", Err: err}
}

func ambiguityCases() []struct {
	name      string
	err       error
	ambiguous bool
} {
	return []struct {
		name      string
		err       error
		ambiguous bool
	}{
		{"deadline exceeded", fmt.Errorf("auth: %w", context.DeadlineExceeded), true},
		{"net timeout", wrapURL(fakeTimeoutErr{}), true},
		{"EOF (response lost)", wrapURL(io.EOF), true},
		{"unexpected EOF", wrapURL(io.ErrUnexpectedEOF), true},
		{"connection reset", wrapURL(&net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("read", syscall.ECONNRESET)}), true},
		{"broken pipe after send", wrapURL(&net.OpError{Op: "write", Net: "tcp", Err: os.NewSyscallError("write", syscall.EPIPE)}), true},
		{"read error after send", wrapURL(&net.OpError{Op: "read", Net: "tcp", Err: errors.New("use of closed network connection")}), true},
		{"connection refused (never sent)", wrapURL(&net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}), false},
		{"explicit 409", &appdto.HTTPError{Code: 409, Message: "conflict"}, false},
		{"explicit 400", &appdto.HTTPError{Code: 400, Message: "bad request"}, false},
		{"explicit 500", &appdto.HTTPError{Code: 500, Message: "internal"}, false},
		{"explicit 503", fmt.Errorf("wrapped: %w", &appdto.HTTPError{Code: 503, Message: "unavailable"}), false},
		{"plain error", errors.New("parse error"), false},
		{"nil", nil, false},
	}
}

func TestIsAmbiguousOutcome_Table(t *testing.T) {
	for _, tc := range ambiguityCases() {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.ambiguous, isAmbiguousOutcome(tc.err))
		})
	}
}

// Step 5 end to end per error type: ambiguous errors query ms-auth by id_user_external before compensating;
// explicit HTTP answers never do.
func TestCreateAuthUser_AmbiguityByErrorType(t *testing.T) {
	for _, tc := range ambiguityCases() {
		if tc.err == nil {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			f := newConfirmFixture(func(f *confirmFixture) {
				f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
					Return(authDto.AuthUserCreateResponseDTO{}, tc.err)
				f.auth.On("ExistsByIDExternal", mock.Anything, "user-123", "test-app", "corr-123").Return(true, nil)
			})
			uc := newConfirmUC(f, RegisterOptions{SagaHardening: true, Metrics: newFakeRegistrationMetrics()})

			_, err := uc.Execute(context.Background(), testConfirmCommand())

			if tc.ambiguous {
				require.NoError(t, err)
				f.auth.AssertCalled(t, "ExistsByIDExternal", mock.Anything, "user-123", "test-app", "corr-123")
				f.auth.AssertNotCalled(t, "HardDeleteUser", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				return
			}
			require.Error(t, err)
			f.auth.AssertNotCalled(t, "ExistsByIDExternal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// Flag off: connection reset is not recovered (previous behavior).
func TestCreateAuthUser_FlagOffResetIsNotRecovered(t *testing.T) {
	f := newConfirmFixture(func(f *confirmFixture) {
		f.auth.On("CreateUser", mock.Anything, mock.Anything, "test-app", "corr-123").
			Return(authDto.AuthUserCreateResponseDTO{}, wrapURL(io.EOF))
	})
	uc := newConfirmUC(f, RegisterOptions{SagaHardening: false})
	_, err := uc.Execute(context.Background(), testConfirmCommand())
	require.Error(t, err)
	f.auth.AssertNotCalled(t, "ExistsByIDExternal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
