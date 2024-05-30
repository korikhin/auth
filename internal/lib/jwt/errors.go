package jwt

import (
	"reflect"

	"github.com/golang-jwt/jwt/v5"
)

func ExpiredOnly(err error) bool {
	target := jwt.ErrTokenExpired
	p := Parent(err, target)

	switch x := p.(type) {
	case interface{ Unwrap() []error }:
		return len(x.Unwrap()) == 1
	case interface{ Unwrap() error }:
		return true
	default:
		return false
	}
}

func Parent(err, target error) error {
	if err == nil || target == nil {
		return nil
	}

	isComparable := reflect.TypeOf(target).Comparable()
	return parent(err, target, isComparable)
}

// NOTE: Overkill for fun
func parent(err, target error, targetComparable bool) error {
	for {
		if targetComparable && err == target {
			return nil
		}
		switch x := err.(type) {
		case interface{ Unwrap() error }:
			p := err
			err = x.Unwrap()
			if err == nil {
				return nil
			}
			if targetComparable && err == target {
				return p
			}
		case interface{ Unwrap() []error }:
			for _, err := range x.Unwrap() {
				if p := parent(err, target, targetComparable); p != nil {
					return p
				}
			}
			return nil
		default:
			return nil
		}
	}
}

// NOTE: Ok solution
// func expiredOnly(err error) bool {
// 	if !errors.Is(err, jwt.ErrTokenExpired) {
// 		return false
// 	}
//
// 	var errs interface{ Unwrap() []error }
// 	if errors.As(err, &errs) {
// 		validationErrs := errs.Unwrap()
// 		if len(validationErrs) > 1 {
// 			var claimErrs interface{ Unwrap() []error }
// 			if errors.As(validationErrs[1], &claimErrs) {
// 				return len(claimErrs.Unwrap()) == 1
// 			}
// 		}
// 	}
//
// 	return false
// }
