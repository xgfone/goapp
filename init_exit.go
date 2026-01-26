package goapp

import "github.com/xgfone/goapp/internal"

func OnInit(f func()) {
	internal.OnInit(f)
}

func OnInitPre(f func()) {
	internal.OnInitPre(f)
}

func OnExit(f func()) {
	internal.OnExit(f)
}

func OnExitPost(f func()) {
	internal.OnExitPost(f)
}
