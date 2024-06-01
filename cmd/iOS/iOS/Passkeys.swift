//
//  Account.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/6/1.
//
import Foundation
import AuthenticationServices
import os

class AccountManager: NSObject, ASAuthorizationControllerPresentationContextProviding, ASAuthorizationControllerDelegate {
	var authenticationAnchor: ASPresentationAnchor?
	private var success: ((Protocols_FinishPasskeysLoginResponse)->Void)?
	private var failure: ((_ error :String) ->Void)?
	private var challenge: Data?
	private var userAgent: String?

	private var loginManager: LoginManager?

	func signInWith(url: String, userAgent: String, success: @escaping (Protocols_FinishPasskeysLoginResponse)->Void, failure: @escaping (_ error:String)->Void) {
		self.authenticationAnchor = ASPresentationAnchor()
		self.success = success
		self.failure = failure
		self.userAgent = userAgent
		
		let parsedURL = URL(string: url)!
		let domain = parsedURL.host()!
		let publicKeyCredentialProvider = ASAuthorizationPlatformPublicKeyCredentialProvider(relyingPartyIdentifier: domain)

		let channel = try! Channel(url: url)
		self.loginManager = LoginManager(channel)
		let challenge = loginManager!.beginLogin()
		self.challenge = challenge

		let assertionRequest = publicKeyCredentialProvider.createCredentialAssertionRequest(challenge: challenge)
		let authController = ASAuthorizationController(authorizationRequests: [ assertionRequest ] )
		authController.delegate = self
		authController.presentationContextProvider = self

		authController.performRequests(options: .preferImmediatelyAvailableCredentials)
	}

	func authorizationController(controller: ASAuthorizationController, didCompleteWithAuthorization authorization: ASAuthorization) {
		switch authorization.credential {
		case let credentialAssertion as ASAuthorizationPlatformPublicKeyCredentialAssertion:
			let rsp = loginManager!.finishLogin(challenge: self.challenge!, userAgent: self.userAgent!, assertion: credentialAssertion)
			print("token:", rsp.token, rsp.cookies)
			
			DispatchQueue.main.async {
				self.success!(rsp)
			}
		default:
			fatalError("Received unknown authorization type.")
		}
	}

	func authorizationController(controller: ASAuthorizationController, didCompleteWithError error: Error) {
		switch error {
		case let credentialError as ASAuthorizationError:
			print(credentialError.localizedDescription)
			self.failure!(credentialError.localizedDescription)
			return
		default:
			fatalError(error.localizedDescription)
		}
	}

	func presentationAnchor(for controller: ASAuthorizationController) -> ASPresentationAnchor {
		return authenticationAnchor!
	}
}
