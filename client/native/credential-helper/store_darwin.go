//go:build darwin

package main

/*
#cgo LDFLAGS: -framework Security -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
#include <stdlib.h>
#include <string.h>

static OSStatus sr_keychain_query(
	const char *service_value,
	const char *account_value,
	const char *keychain_path,
	CFMutableDictionaryRef *query_output
) {
	CFStringRef service = CFStringCreateWithCString(NULL, service_value, kCFStringEncodingUTF8);
	CFStringRef account = CFStringCreateWithCString(NULL, account_value, kCFStringEncodingUTF8);
	if (service == NULL || account == NULL) {
		if (service != NULL) CFRelease(service);
		if (account != NULL) CFRelease(account);
		return errSecAllocate;
	}

	CFMutableDictionaryRef query = CFDictionaryCreateMutable(
		NULL,
		0,
		&kCFTypeDictionaryKeyCallBacks,
		&kCFTypeDictionaryValueCallBacks
	);
	if (query == NULL) {
		CFRelease(service);
		CFRelease(account);
		return errSecAllocate;
	}

	CFDictionarySetValue(query, kSecClass, kSecClassGenericPassword);
	CFDictionarySetValue(query, kSecAttrService, service);
	CFDictionarySetValue(query, kSecAttrAccount, account);
	CFRelease(service);
	CFRelease(account);

	if (keychain_path != NULL && keychain_path[0] != '\0') {
		SecKeychainRef keychain = NULL;
		OSStatus status = SecKeychainOpen(keychain_path, &keychain);
		if (status != errSecSuccess) {
			CFRelease(query);
			return status;
		}
		const void *keychains[] = { keychain };
		CFArrayRef search_list = CFArrayCreate(
			NULL,
			keychains,
			1,
			&kCFTypeArrayCallBacks
		);
		if (search_list == NULL) {
			CFRelease(keychain);
			CFRelease(query);
			return errSecAllocate;
		}
		CFDictionarySetValue(query, kSecUseKeychain, keychain);
		CFDictionarySetValue(query, kSecMatchSearchList, search_list);
		CFRelease(search_list);
		CFRelease(keychain);
	}

	*query_output = query;
	return errSecSuccess;
}

static OSStatus sr_keychain_load(
	const char *service,
	const char *account,
	const char *keychain_path,
	unsigned char **output,
	long *output_length
) {
	CFMutableDictionaryRef query = NULL;
	OSStatus status = sr_keychain_query(service, account, keychain_path, &query);
	if (status != errSecSuccess) return status;
	CFDictionarySetValue(query, kSecReturnData, kCFBooleanTrue);
	CFDictionarySetValue(query, kSecMatchLimit, kSecMatchLimitOne);

	CFTypeRef result = NULL;
	status = SecItemCopyMatching(query, &result);
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
	const char *keychain_path,
	const unsigned char *secret,
	long secret_length
) {
	CFMutableDictionaryRef query = NULL;
	OSStatus status = sr_keychain_query(service, account, keychain_path, &query);
	if (status != errSecSuccess) return status;
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
	status = SecItemUpdate(query, updates);
	if (status == errSecItemNotFound) {
		CFDictionaryRemoveValue(query, kSecMatchSearchList);
		CFDictionarySetValue(query, kSecValueData, data);
		status = SecItemAdd(query, NULL);
	}

	CFRelease(updates);
	CFRelease(data);
	CFRelease(query);
	return status;
}

static OSStatus sr_keychain_clear(
	const char *service,
	const char *account,
	const char *keychain_path
) {
	CFMutableDictionaryRef query = NULL;
	OSStatus status = sr_keychain_query(service, account, keychain_path, &query);
	if (status != errSecSuccess) return status;
	status = SecItemDelete(query);
	CFRelease(query);
	return status;
}
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

const (
	errSecSuccess      = 0
	errSecItemNotFound = -25300
)

const credentialKeychainPathEnvironment = "SPACE_ROCKS_CREDENTIAL_KEYCHAIN_PATH"

type darwinCredentialStore struct {
	keychainPath string
}

func newPlatformStore() credentialStore {
	return darwinCredentialStore{
		keychainPath: os.Getenv(credentialKeychainPathEnvironment),
	}
}

func (store darwinCredentialStore) Load(req request) (string, error) {
	service := C.CString(req.Service)
	account := C.CString(req.Account)
	keychainPath := C.CString(store.keychainPath)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	defer C.free(unsafe.Pointer(keychainPath))

	var output *C.uchar
	var outputLength C.long
	status := int(C.sr_keychain_load(
		service,
		account,
		keychainPath,
		&output,
		&outputLength,
	))
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

func (store darwinCredentialStore) Save(req request) error {
	service := C.CString(req.Service)
	account := C.CString(req.Account)
	keychainPath := C.CString(store.keychainPath)
	secret := C.CBytes([]byte(req.Secret))
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	defer C.free(unsafe.Pointer(keychainPath))
	defer C.free(secret)

	status := int(C.sr_keychain_save(
		service,
		account,
		keychainPath,
		(*C.uchar)(secret),
		C.long(len(req.Secret)),
	))
	if status != errSecSuccess {
		return fmt.Errorf("SecItemAdd or SecItemUpdate failed with OSStatus %d", status)
	}
	return nil
}

func (store darwinCredentialStore) Clear(req request) error {
	service := C.CString(req.Service)
	account := C.CString(req.Account)
	keychainPath := C.CString(store.keychainPath)
	defer C.free(unsafe.Pointer(service))
	defer C.free(unsafe.Pointer(account))
	defer C.free(unsafe.Pointer(keychainPath))

	status := int(C.sr_keychain_clear(service, account, keychainPath))
	if status == errSecItemNotFound {
		return errCredentialNotFound
	}
	if status != errSecSuccess {
		return fmt.Errorf("SecItemDelete failed with OSStatus %d", status)
	}
	return nil
}
