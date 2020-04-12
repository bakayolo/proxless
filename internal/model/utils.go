package model

import "kube-proxless/internal/utils"

func compareRoute(r1, r2 *Route) bool {
	if r1.service == r2.service &&
		r1.port == r2.port &&
		r1.namespace == r2.namespace &&
		r1.deployment == r2.deployment &&
		utils.CompareUnorderedArray(r1.domains, r2.domains) {
		return true
	}

	return false
}
