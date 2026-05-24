package domain

import "errors"

func CanTransition(from, to string) error {

	switch from {

	case "PENDING":
		if to == "PAID" || to == "CANCELLED" {
			return nil
		}

	case "PAID":
		if to == "PROCESSING" || to == "CANCELLED" {
			return nil
		}

	case "PROCESSING":
		if to == "SHIPPED" {
			return nil
		}

	case "SHIPPED":
		if to == "DELIVERED" {
			return nil
		}
	}

	return errors.New("invalid state transition")
}
