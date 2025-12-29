//go:build darwin

// Package monitor provides platform-specific sleep and network monitoring hooks.
// It is used by the agent and can be reused by client code.

package monitor

/*
#cgo LDFLAGS: -framework SystemConfiguration -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <SystemConfiguration/SystemConfiguration.h>

static CFRunLoopRef netRunLoopRef = NULL;
static SCDynamicStoreRef netStoreRef = NULL;

extern void networkCallbackGo();

static void networkCallback(SCDynamicStoreRef store, CFArrayRef changedKeys, void *info) {
	(void)store;
	(void)changedKeys;
	(void)info;
	networkCallbackGo();
}

static int startNetworkMonitor() {
	SCDynamicStoreContext ctx = {0, NULL, NULL, NULL, NULL};
	netStoreRef = SCDynamicStoreCreate(NULL, CFSTR("rpa-network"), networkCallback, &ctx);
	if (netStoreRef == NULL) {
		return 1;
	}
	CFStringRef keys[2];
	keys[0] = CFSTR("State:/Network/Global/IPv4");
	keys[1] = CFSTR("State:/Network/Global/IPv6");
	CFArrayRef keyArray = CFArrayCreate(NULL, (const void **)keys, 2, &kCFTypeArrayCallBacks);
	if (keyArray == NULL) {
		return 2;
	}
	Boolean ok = SCDynamicStoreSetNotificationKeys(netStoreRef, keyArray, NULL);
	CFRelease(keyArray);
	if (!ok) {
		return 3;
	}
	CFRunLoopSourceRef src = SCDynamicStoreCreateRunLoopSource(NULL, netStoreRef, 0);
	if (src == NULL) {
		return 4;
	}
	netRunLoopRef = CFRunLoopGetCurrent();
	CFRunLoopAddSource(netRunLoopRef, src, kCFRunLoopDefaultMode);
	CFRelease(src);
	CFRunLoopRun();
	return 0;
}

static void stopNetworkMonitor() {
	if (netRunLoopRef != NULL) {
		CFRunLoopStop(netRunLoopRef);
	}
	if (netStoreRef != NULL) {
		CFRelease(netStoreRef);
		netStoreRef = NULL;
	}
	netRunLoopRef = NULL;
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
	networkEventMu sync.Mutex
	networkEventCh chan struct{}
)

//export networkCallbackGo
func networkCallbackGo() {
	networkEventMu.Lock()
	ch := networkEventCh
	networkEventMu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- struct{}{}:
	default:
	}
}

func StartNetworkMonitor(ctx context.Context, _ config.AgentConfig, logger *logging.Logger, onEvent func(reason string)) {
	if onEvent == nil {
		onEvent = func(string) {}
	}
	logger.Info("network monitor: using SystemConfiguration")
	ch := make(chan struct{}, 8)
	networkEventMu.Lock()
	networkEventCh = ch
	networkEventMu.Unlock()
	defer func() {
		networkEventMu.Lock()
		networkEventCh = nil
		networkEventMu.Unlock()
	}()

	go func() {
		rc := int(C.startNetworkMonitor())
		if rc != 0 {
			logger.Error("network monitor failed to start (rc=%d)", rc)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			C.stopNetworkMonitor()
			return
		case <-ch:
			onEvent("network change")
		}
	}
}
