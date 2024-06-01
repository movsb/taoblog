//
//  TaoBlogApp.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/5/31.
//

import SwiftUI

class GlobalState : ObservableObject {
	@Published var signedIn = false
	@Published var token = ""
	@Published var cookies: [Protocols_FinishPasskeysLoginResponse.Cookie] = []
	@Published var url = ""

	public var userAgent = "taoblog-ios-client/1.0"
	
	private var channel: Channel?
	public var client: Client?
	
	public func login(rsp: Protocols_FinishPasskeysLoginResponse, url: String) {
		self.url = url
		self.cookies = rsp.cookies
		self.token = rsp.token
		
		let channel = try! Channel(url: url)
		let client = Client(channel: channel, token: rsp.token)
		
		self.channel = channel
		self.client = client
		self.signedIn = true
	}
}

@main
struct TaoBlogApp: App {
	@StateObject var globalStates = GlobalState()
	
    var body: some Scene {
        WindowGroup {
			ContentView()
				.environmentObject(globalStates)
        }
    }
}
