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

public class Client {
	private var _eventLoopGroup: MultiThreadedEventLoopGroup
	private var _grpcChannel: GRPCChannel

	private var _blog: Protocols_TaoBlogNIOClient
	
	public init() {
		self._eventLoopGroup = MultiThreadedEventLoopGroup(numberOfThreads: 1)

		self._grpcChannel = try! GRPCChannelPool.with(
			target: .host("192.168.1.102", port: 2563),
			transportSecurity: .plaintext,
			eventLoopGroup: self._eventLoopGroup)
		
		self._blog = Protocols_TaoBlogNIOClient(
			channel: self._grpcChannel,
			defaultCallOptions: .init(
				customMetadata: .init([
					("token", "12345678")
				])
			)
		)
	}
	
	deinit {
		try! self._grpcChannel.close().wait()
		
		try! self._eventLoopGroup.syncShutdownGracefully()
	}

	func listPosts(completion: @escaping ([Protocols_Post])->Void) throws {
		self._blog.listPosts(.with {
			$0.kinds = ["post", "tweet"]
			$0.orderBy = "date desc"
		}).response.whenSuccess { rsp in
			DispatchQueue.main.async {
				completion(rsp.posts)
			}
		}
	}
}
