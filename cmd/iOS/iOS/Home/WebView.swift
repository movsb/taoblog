//
//  WebView.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/6/1.
//

import SwiftUI
import WebKit

struct WebView : View {
	@EnvironmentObject var states: GlobalState
	
	private let url: String
	@State private var _alerting = false
	@State private var _alertingMessage = ""
	@State private var confirmed: ((Bool)->Void)?
	
	init(url: String) {
		self.url = url
	}
	
	var body: some View {
		_WebView(url: url, states: states, parent: self)
			.alert(isPresented: $_alerting) {
				Alert(
					title: Text(_alertingMessage),
					message: Text(""),
					primaryButton: .default(
						Text("取消"),
						action: {
							confirmed!(false)
						}
					),
					secondaryButton: .destructive(
						Text("确认"),
						action: {
							confirmed!(true)
						}
					)
				)
			}
	}
	
	public func confirm(message: String, confirmed: @escaping (Bool)->Void) {
		self.confirmed = confirmed
		_alertingMessage = message
		_alerting = true
	}
}

struct _WebView: UIViewRepresentable {
	// 带不过来，手动传了。
	private var states: GlobalState
	
	private let url: String
	private let webView: WKWebView
	private let parent: WebView
	
	class Delegate : NSObject, WKUIDelegate {
		let parent: _WebView
		init(_ parent: _WebView) {
			self.parent = parent
		}
		func webView(_ webView: WKWebView, runJavaScriptConfirmPanelWithMessage message: String, initiatedByFrame frame: WKFrameInfo, completionHandler: @escaping (Bool) -> Void) {
			print(message)
			parent.parent.confirm(message: message) {
				completionHandler($0)
			}
		}
		func webView(_ webView: WKWebView, runJavaScriptAlertPanelWithMessage message: String, initiatedByFrame frame: WKFrameInfo) async {
			print(message)
		}
	}
	
	init(url: String, states: GlobalState, parent: WebView) {
		self.states = states
		self.url = url
		self.parent = parent
		webView = WKWebView()
		webView.allowsBackForwardNavigationGestures = true
		webView.customUserAgent = states.userAgent
		
		let cookieStore = webView.configuration.websiteDataStore.httpCookieStore
		states.cookies.forEach { c in
			let cookie = HTTPCookie(properties: [
				.domain: URL(string: url)!.host()!,
				.path: "/",
				.name: c.name,
				.value: c.value,
				.secure: true,
				.sameSitePolicy: "lax",
				// httpOnly 去哪儿了？
			])
			print("设置 Cookie：", cookie!)
			cookieStore.setCookie(cookie!)
		}
	}
	
	func makeCoordinator() -> Delegate {
		return Delegate(self)
	}
	
	func makeUIView(context: Context) -> WKWebView {
		return webView
	}
	
	func updateUIView(_ uiView: WKWebView, context: Context) {
		webView.uiDelegate = context.coordinator
		webView.load(URLRequest(url: URL(string: url)!))
	}
}
