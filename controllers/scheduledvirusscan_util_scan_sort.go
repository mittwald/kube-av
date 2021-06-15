package controllers

import avv1beta1 "github.com/mittwald/kube-av/api/v1beta1"

type scanList []avv1beta1.VirusScan

func (s scanList) Len() int {
	return len(s)
}

func (s scanList) Less(i, j int) bool {
	return s[i].CreationTimestamp.Unix() < s[j].CreationTimestamp.Unix()
}

func (s scanList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
