//
//  LoginView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/6/1.
//

import SwiftUI

struct Portal : Identifiable {
	var id: String // url
	var rsp: Protocols_FinishPasskeysLoginResponse?

	public init(id: String) {
		self.init(id: id, rsp: nil)
	}
	public init(id: String, rsp: Protocols_FinishPasskeysLoginResponse?) {
		self.id = id
		self.rsp = rsp
	}
}

var portals = [
	Portal(id: "https://blog.twofei.com"),
	Portal(
		id: "https://blog.mac.twofei.com",
		rsp: Protocols_FinishPasskeysLoginResponse.with {
			$0.token = "2:12345678"
			$0.cookies = [
				Protocols_FinishPasskeysLoginResponse.Cookie.with {
					$0.name = "taoblog.login"
					$0.value = "6e60db3f03c45305613870f9bbe7da5321bc2e8b"
					$0.httpOnly = true
				},
				Protocols_FinishPasskeysLoginResponse.Cookie.with {
					$0.name = "taoblog.user_id"
					$0.value = "2"
					$0.httpOnly = false
				}
			]
		}
	),
]

var accountManager = AccountManager()

struct LoginView: View {
	@EnvironmentObject var states : GlobalState
	@State private var _alerting = false
	@State private var _alertMessage = ""
	var body: some View {
		VStack {
			Text("陪她去流浪")
				.font(.largeTitle)
				.bold()
				.padding(.bottom)
			ForEach(portals) { p in
				Button {
					if let rsp = p.rsp {
						states.login(rsp: rsp, url: p.id)
					} else {
						accountManager.signInWith(url: p.id, userAgent: states.userAgent, success: {rsp in
							states.login(rsp: rsp, url: p.id)
							states.signedIn = true
							print("登录成功.")
						}, failure: { error in
							_alertMessage = error
							_alerting = true
						})
					}
				} label: {
					VStack {
						Label("通行密钥登录", systemImage: "key")
							.font(.headline)
						Text(p.id)
							.font(.caption)
					}
				}
				.padding(.bottom)
			}
		}
		.alert(isPresented: $_alerting) {
			Alert(title: Text(_alertMessage))
		}
	}
}

#Preview {
	LoginView()
}
