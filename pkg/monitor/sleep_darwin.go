//go:build darwin

// Package monitor provides platform-specific sleep and network monitoring hooks.
// It is used by the agent and can be reused by client code.

package monitor

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/IOMessage.h>
#include <IOKit/pwr_mgt/IOPMLib.h>

static IONotificationPortRef notifyPortRef = NULL;
static io_object_t notifierObject = 0;
static io_connect_t rootPort = 0;
static CFRunLoopRef runLoopRef = NULL;

extern void powerCallbackGo(uint32_t messageType);

static void powerCallback(void *refCon, io_service_t service, natural_t messageType, void *messageArgument) {
	switch (messageType) {
		case kIOMessageCanSystemSleep:
			IOAllowPowerChange(rootPort, (long)messageArgument);
			break;
		case kIOMessageSystemWillSleep:
			IOAllowPowerChange(rootPort, (long)messageArgument);
			powerCallbackGo((uint32_t)messageType);
			break;
		case kIOMessageSystemHasPoweredOn:
			powerCallbackGo((uint32_t)messageType);
			break;
		default:
			break;
	}
}

static int startPowerMonitor() {
	rootPort = IORegisterForSystemPower(NULL, &notifyPortRef, powerCallback, &notifierObject);
	if (rootPort == 0) {
		return 1;
	}
	runLoopRef = CFRunLoopGetCurrent();
	CFRunLoopAddSource(runLoopRef, IONotificationPortGetRunLoopSource(notifyPortRef), kCFRunLoopDefaultMode);
	CFRunLoopRun();
	return 0;
}

static void stopPowerMonitor() {
	if (runLoopRef != NULL) {
		CFRunLoopStop(runLoopRef);
	}
	if (notifierObject != 0) {
		IOObjectRelease(notifierObject);
		notifierObject = 0;
	}
	if (notifyPortRef != NULL) {
		IONotificationPortDestroy(notifyPortRef);
		notifyPortRef = NULL;
	}
	if (rootPort != 0) {
		IOServiceClose(rootPort);
		rootPort = 0;
	}
	runLoopRef = NULL;
}
*/
import "C"

import (
	"context"
	"sync"

	"reverse-proxy-agent/pkg/config"
	"reverse-proxy-agent/pkg/logging"
)

var (
	sleepEventMu sync.Mutex
	sleepEventCh chan uint32
)

//export powerCallbackGo
func powerCallbackGo(messageType C.uint32_t) {
	sleepEventMu.Lock()
	ch := sleepEventCh
	sleepEventMu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- uint32(messageType):
	default:
	}
}

func StartSleepMonitor(ctx context.Context, cfg config.AgentConfig, logger *logging.Logger, onEvent func(reason string)) {
	if cfg.SleepCheckSec <= 0 {
		return
	}
	if onEvent == nil {
		onEvent = func(string) {}
	}

	logger.Info("sleep monitor: using IOKit")
	ch := make(chan uint32, 8)

	sleepEventMu.Lock()
	sleepEventCh = ch
	sleepEventMu.Unlock()
	defer func() {
		sleepEventMu.Lock()
		sleepEventCh = nil
		sleepEventMu.Unlock()
	}()

	go func() {
		rc := int(C.startPowerMonitor())
		if rc != 0 {
			logger.Error("IOKit power monitor failed to start (rc=%d)", rc)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			C.stopPowerMonitor()
			return
		case msg := <-ch:
			switch msg {
			case uint32(C.kIOMessageSystemWillSleep):
				logger.Info("sleep detected")
				onEvent("sleep")
			case uint32(C.kIOMessageSystemHasPoweredOn):
				logger.Info("wake detected")
				onEvent("wake")
			}
		}
	}
}
