//
//  Client.swift
//  TaoBlog
//
//  Created by Yang Tao on 2024/5/31.
//

import Foundation
import Foundation
import GRPC
import NIOCore
import NIOPosix
import AuthenticationServices

extension String: LocalizedError {
	public var errorDescription: String? { return self }
}

class Channel {
	private var _eventLoopGroup: MultiThreadedEventLoopGroup
	private var _grpcChannel: GRPCChannel

	public init(url: String) throws {
		self._eventLoopGroup = MultiThreadedEventLoopGroup(numberOfThreads: 1)
		
		guard let u = URL(string: url) else { throw "地址不正确。" }
		let host = u.host()!
		var port = u.port
		if port == nil {
			port = 443
		}

		self._grpcChannel = try! GRPCChannelPool.with(
			target: .host(host, port: port!),
			transportSecurity: .tls(.makeClientDefault(compatibleWith: self._eventLoopGroup)),
			eventLoopGroup: self._eventLoopGroup)
	}
	
	public func channel() -> GRPCChannel {
		return _grpcChannel
	}
	
	deinit {
		try! self._grpcChannel.close().wait()
		
		try! self._eventLoopGroup.syncShutdownGracefully()
	}
}

class LoginManager {
	private var channel: Channel
	private var client: Protocols_AuthNIOClient
	
	public init(_ channel: Channel) {
		self.channel = channel
		self.client = Protocols_AuthNIOClient(channel: channel.channel())
	}
	
	// 开始登录，返回 Challenge。
	public func beginLogin() -> Data {
		return try! client.beginPasskeysLogin(.init()).response.wait().challenge
	}

	public func finishLogin(challenge: Data, userAgent: String, assertion: ASAuthorizationPlatformPublicKeyCredentialAssertion) -> Protocols_FinishPasskeysLoginResponse {
		let req = Protocols_FinishPasskeysLoginRequest.with {
			$0.attachment = Int32(assertion.attachment.rawValue)
			$0.authenticatorData = assertion.rawAuthenticatorData
			$0.challenge = challenge
			$0.clientDataJson = assertion.rawClientDataJSON.base64EncodedString()
			$0.id = assertion.credentialID
			$0.signature = assertion.signature
			$0.userID = assertion.userID
			$0.userAgent = userAgent
		}
		return try! client.finishPasskeysLogin(req).response.wait()
	}
}

public class Client : ObservableObject {
	private var _blog: Protocols_TaoBlogNIOClient
	private var _auth: Protocols_AuthNIOClient
	
	init(channel: Channel, token:String) {
		self._blog = Protocols_TaoBlogNIOClient(
			channel: channel.channel(),
			defaultCallOptions: .init(
				customMetadata: .init([
					("Authorization", "token \(token)")
				])
			)
		)
		self._auth = Protocols_AuthNIOClient(
			channel: channel.channel(),
			defaultCallOptions: .init(
				customMetadata: .init([
					("Authorization", "token \(token)")
				])
			)
		)
	}
	
	func getInfo(completion: @escaping (Protocols_GetInfoResponse)->Void) {
		self._blog.getInfo(.init()).response.whenComplete { result in
			switch result {
			case .success(let success):
				print(success)
				DispatchQueue.main.async {
					completion(success)
				}
			case .failure(let failure):
				print(failure)
				fatalError(failure.localizedDescription)
			}
		}
	}

	func createPost(_ post: Protocols_Post, completion: @escaping (Result<Protocols_Post, any Error>)->Void) {
		self._blog.createPost(post).response.whenComplete { result in
				switch result {
				case .success(let success):
					print(success)
				case .failure(let failure):
					print(failure)
				}
			DispatchQueue.main.async {
				completion(result)
			}
		}
	}
	
	func updatePost(_ post: Protocols_Post, completion: @escaping (Result<Protocols_Post, any Error>)->Void) {
		self._blog.updatePost(Protocols_UpdatePostRequest.with {
			$0.post = post
			$0.updateMask = .with {
				$0.paths = ["source", "source_type"]
			}
		}).response.whenComplete { result in
				switch result {
				case .success(let success):
					print(success)
				case .failure(let failure):
					print(failure)
				}
			DispatchQueue.main.async {
				completion(result)
			}
		}
	}
	
	func listPosts(completion: @escaping ([Protocols_Post])->Void) throws {
		self._blog.listPosts(.with {
			$0.kinds = ["post", "tweet"]
			$0.orderBy = "date desc"
			$0.limit = 1000
		}).response.whenComplete { result in
			switch result {
			case .success(let success):
				DispatchQueue.main.async {
					completion(success.posts)
				}
			case .failure(let failure):
				fatalError(failure.localizedDescription)
			}
		}
	}

	func setPostStatus(_ id: Int64, status: String, completion: @escaping ()->Void) throws {
		self._blog.setPostStatus(.with {
			$0.id = id
			$0.status = status
			$0.touch = false
		}).response.whenComplete { result in
			switch result {
			case .success(let success):
				DispatchQueue.main.async {
					completion()
				}
			case .failure(let failure):
				fatalError(failure.localizedDescription)
			}
		}
	}
}
