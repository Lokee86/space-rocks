//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Security -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

static CFMutableDictionaryRef sr_keychain_query(const char *service_value, const char *account_value) {
	CFStringRef service = CFStringCreateWithCString(NULL, service_value, kCFStringEncodingUTF8);
	CFStringRef account = CFStringCreateWithCString(NULL, account_value, kCFStringEncodingUTF8);
	if (service == NULL || account == NULL) {
		if (service != NULL) CFRelease(service);
		if (account != NULL) CFRelease(account);
		return NULL;
	}

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		NULL,
		0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);
	if (query != NULL) {
		CFDictionarySetValue(query, kSecClass, kSecClassGenericPassword);
		CFDictionarySetValue(query, kSecAttrService, service);
		CFDictionarySetValue(query, kSecAttrAccount, account);
	}
	CFRelease(service);
	CFRelease(account);
	return query;
}

static OSStatus sr_keychain_load(
	const char *service,
	const char *account,
	unsigned char **output,
	long *output_length
) {
	CFMutableDictionaryRef query = sr_keychain_query(service, account);
	if (query == NULL) return errSecAllocate;
	CFDictionarySetValue(query, kSecReturnData, kCFBooleanTrue);
	CFDictionarySetValue(query, kSecMatchLimit, kSecMatchLimitOne);

	CFTypeRef result = NULL;
	OSStatus status = SecItemCopyMatching(query, &result);
	CFRelease(query);
	if (status != errSecSuccess) return status;
	if (result == NULL || CFGetTypeID(result) != CFDataGetTypeID()) {
		if (result != NULL) CFRelease(result);
		return errSecDecode;
	}

	CFDataRef data = (CFDataRef)result;
	CFIndex length = CFDataGetLength(data);
	unsigned char *copy = NULL;
	if (length > 0) {
		copy = (unsigned char *)malloc((size_t)length);
		if (copy == NULL) {
			CFRelease(result);
			return errSecAllocate;
		}
		memcpy(copy, CFDataGetBytePtr(data), (size_t)length);
	}
	CFRelease(result);
	*output = copy;
	*output_length = (long)length;
	return errSecSuccess;
}

static OSStatus sr_keychain_save(
	const char *service,
	const char *account,
	const unsigned char *secret,
	long secret_length
) {
	CFMutableDictionaryRef query = sr_keychain_query(service, account);
	if (query == NULL) return errSecAllocate;
	CFDataRef data = CFDataCreate(NULL, secret, (CFIndex)secret_length);
	if (data == NULL) {
		CFRelease(query);
		return errSecAllocate;
	}

	const void *keys[] = { kSecValueData };
	const void *values[] = { data };
	CFDictionaryRef updates = CFDictionaryCreate(
		NULL,
		keys,
		values,
		1,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);
	if (updates == NULL) {
		CFRelease(data);
		CFRelease(query);
		return errSecAllocate;
	}
	OSStatus status = SecItemUpdate(query, updates);
	if (status == errSecItemNotFound) {
		CFDictionarySetValue(query, kSecValueData, data);
		status = SecItemAdd(query, NULL);
	}

	CFRelease(updates);
	CFRelease(data);
	CFRelease(query);
	return status;
}

static OSStatus sr_keychain_clear(const char *service, const char *account) {
	CFMutableDictionaryRef query = sr_keychain_query(service, account);
	if (query == NULL) return errSecAllocate;
	OSStatus status = SecItemDelete(query);
	CFRelease(query);
	return status;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	errSecSuccess      = 0
	errSecItemNotFound = -25300
)

type darwinCredentialStore struct{}

func newPlatformStore() credentialStore {
	return darwinCredentialStore{}
}

func (darwinCredentialStore) Load(req request) (string, error) {
	service := C.CString(req.Service)
	account := C.CString(req.Account)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))

	var output *C.uchar
	var outputLength C.long
	status := int(C.sr_keychain_load(service, account, &output, &outputLength))
	if status == errSecItemNotFound {
		return "", errCredentialNotFound
	}
	if status != errSecSuccess {
		return "", fmt.Errorf("SecItemCopyMatching failed with OSStatus %d", status)
	}
	if output == nil || outputLength == 0 {
		return "", nil
	}
	defer C.free(unsafe.Pointer(output))
	return string(C.GoBytes(unsafe.Pointer(output), C.int(outputLength))), nil
}

func (darwinCredentialStore) Save(req request) error {
	service := C.CString(req.Service)
	account := C.CString(req.Account)
	secret := C.CBytes([]byte(req.Secret))
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	defer C.free(secret)

	status := int(C.sr_keychain_save(
		service,
		account,
		(*C.uchar)(secret),
		C.long(len(req.Secret)),
	))
	if status != errSecSuccess {
		return fmt.Errorf("SecItemAdd or SecItemUpdate failed with OSStatus %d", status)
	}
	return nil
}

func (darwinCredentialStore) Clear(req request) error {
	service := C.CString(req.Service)
	account := C.CString(req.Account)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))

	status := int(C.sr_keychain_clear(service, account))
	if status == errSecItemNotFound {
		return errCredentialNotFound
	}
	if status != errSecSuccess {
		return fmt.Errorf("SecItemDelete failed with OSStatus %d", status)
	}
	return nil
}
