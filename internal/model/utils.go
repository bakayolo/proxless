package model

import "kube-proxless/internal/utils"

func (r *Route) isEqual(r2 *Route) bool {
	if r.service == r2.service &&
		r.port == r2.port &&
		r.namespace == r2.namespace &&
		r.deployment == r2.deployment &&
		utils.CompareUnorderedArray(r.domains, r2.domains) {
		return true
	}

	return false
}
