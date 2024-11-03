//go:build windows

package mi

import (
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	applicationID = "windows_exporter"

	LocaleEnglish = "en-us"
)

var (
	// DestinationOptionsTimeout is the key for the timeout option.
	//
	// https://github.com/microsoft/win32metadata/blob/527806d20d83d3abd43d16cd3fa8795d8deba343/generation/WinSDK/RecompiledIdlHeaders/um/mi.h#L7830
	DestinationOptionsTimeout = UTF16PtrFromString[*uint16]("__MI_DESTINATIONOPTIONS_TIMEOUT")

	// DestinationOptionsUILocale is the key for the UI locale option.
	//
	// https://github.com/microsoft/win32metadata/blob/527806d20d83d3abd43d16cd3fa8795d8deba343/generation/WinSDK/RecompiledIdlHeaders/um/mi.h#L8248
	DestinationOptionsUILocale = UTF16PtrFromString[*uint16]("__MI_DESTINATIONOPTIONS_UI_LOCALE")
)

var (
	modMi = windows.NewLazySystemDLL("mi.dll")

	procMIApplicationInitialize = modMi.NewProc("MI_Application_InitializeV1")
)

// Application represents the MI application.
// https://learn.microsoft.com/de-de/windows/win32/api/mi/ns-mi-mi_application
type Application struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *ApplicationFT
}

// ApplicationFT represents the function table of the MI application.
// https://learn.microsoft.com/de-de/windows/win32/api/mi/ns-mi-mi_applicationft
type ApplicationFT struct {
	Close                          uintptr
	NewSession                     uintptr
	NewHostedProvider              uintptr
	NewInstance                    uintptr
	NewDestinationOptions          uintptr
	NewOperationOptions            uintptr
	NewSubscriptionDeliveryOptions uintptr
	NewSerializer                  uintptr
	NewDeserializer                uintptr
	NewInstanceFromClass           uintptr
	NewClass                       uintptr
}

type DestinationOptions struct {
	reserved1 uint64
	reserved2 uintptr
	ft        *DestinationOptionsFT
}

type DestinationOptionsFT struct {
	Delete                   uintptr
	SetString                uintptr
	SetNumber                uintptr
	AddCredentials           uintptr
	GetString                uintptr
	GetNumber                uintptr
	GetOptionCount           uintptr
	GetOptionAt              uintptr
	GetOption                uintptr
	GetCredentialsCount      uintptr
	GetCredentialsAt         uintptr
	GetCredentialsPasswordAt uintptr
	Clone                    uintptr
	SetInterval              uintptr
	GetInterval              uintptr
}

// Application_Initialize initializes the MI [Application].
// It is recommended to have only one Application per process.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_initializev1
func Application_Initialize() (*Application, error) {
	application := &Application{}

	applicationId, err := windows.UTF16PtrFromString(applicationID)
	if err != nil {
		return nil, err
	}

	r0, _, err := procMIApplicationInitialize.Call(
		0,
		uintptr(unsafe.Pointer(applicationId)),
		0,
		uintptr(unsafe.Pointer(application)),
	)

	if !errors.Is(err, windows.NOERROR) {
		return nil, fmt.Errorf("syscall returned: %w", err)
	}

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	return application, nil
}

// Close deinitializes the management infrastructure client API that was initialized through a call to Application_Initialize.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_close
func (application *Application) Close() error {
	if application == nil || application.ft == nil {
		return ErrNotInitialized
	}

	r0, _, _ := syscall.SyscallN(application.ft.Close, uintptr(unsafe.Pointer(application)))

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

// NewSession creates a session used to share connections for a set of operations to a single destination.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_newsession
func (application *Application) NewSession(options *DestinationOptions) (*Session, error) {
	if application == nil || application.ft == nil {
		return nil, ErrNotInitialized
	}

	session := &Session{}

	r0, _, _ := syscall.SyscallN(
		application.ft.NewSession,
		uintptr(unsafe.Pointer(application)),
		0,
		0,
		uintptr(unsafe.Pointer(options)),
		0,
		0,
		uintptr(unsafe.Pointer(session)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	defaultOperationOptions, err := application.NewOperationOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create default operation options: %w", err)
	}

	if err = defaultOperationOptions.SetTimeout(5 * time.Second); err != nil {
		return nil, fmt.Errorf("failed to set timeout: %w", err)
	}

	session.defaultOperationOptions = defaultOperationOptions

	return session, nil
}

// NewOperationOptions creates an OperationOptions object that can be used with the operation functions on the Session object.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_newoperationoptions
func (application *Application) NewOperationOptions() (*OperationOptions, error) {
	if application == nil || application.ft == nil {
		return nil, ErrNotInitialized
	}

	operationOptions := &OperationOptions{}
	mustUnderstand := True

	r0, _, _ := syscall.SyscallN(
		application.ft.NewOperationOptions,
		uintptr(unsafe.Pointer(application)),
		uintptr(mustUnderstand),
		uintptr(unsafe.Pointer(operationOptions)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	return operationOptions, nil
}

// NewDestinationOptions creates an DestinationOptions object that can be used with the Application.NewSession function.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_application_newdestinationoptions
func (application *Application) NewDestinationOptions() (*DestinationOptions, error) {
	if application == nil || application.ft == nil {
		return nil, ErrNotInitialized
	}

	operationOptions := &DestinationOptions{}

	r0, _, _ := syscall.SyscallN(
		application.ft.NewDestinationOptions,
		uintptr(unsafe.Pointer(application)),
		uintptr(unsafe.Pointer(operationOptions)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return nil, result
	}

	return operationOptions, nil
}

// SetTimeout sets the timeout for the destination options.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_destinationoptions_settimeout
func (do *DestinationOptions) SetTimeout(timeout time.Duration) error {
	if do == nil || do.ft == nil {
		return ErrNotInitialized
	}

	r0, _, _ := syscall.SyscallN(
		do.ft.SetInterval,
		uintptr(unsafe.Pointer(do)),
		uintptr(unsafe.Pointer(DestinationOptionsTimeout)),
		uintptr(unsafe.Pointer(NewInterval(timeout))),
		0,
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

// SetLocale sets the locale for the destination options.
//
// https://learn.microsoft.com/en-us/windows/win32/api/mi/nf-mi-mi_destinationoptions_setuilocale
func (do *DestinationOptions) SetLocale(locale string) error {
	if do == nil || do.ft == nil {
		return ErrNotInitialized
	}

	localeUTF16, err := windows.UTF16PtrFromString(locale)
	if err != nil {
		return fmt.Errorf("failed to convert locale: %w", err)
	}

	r0, _, _ := syscall.SyscallN(
		do.ft.SetString,
		uintptr(unsafe.Pointer(do)),
		uintptr(unsafe.Pointer(DestinationOptionsUILocale)),
		uintptr(unsafe.Pointer(localeUTF16)),
		0,
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}

func (do *DestinationOptions) Delete() error {
	r0, _, _ := syscall.SyscallN(
		do.ft.Delete,
		uintptr(unsafe.Pointer(do)),
	)

	if result := ResultError(r0); !errors.Is(result, MI_RESULT_OK) {
		return result
	}

	return nil
}
