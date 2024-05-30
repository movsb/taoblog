//
//  main.swift
//  ocr
//
//  Created by Yang Tao on 2024/5/31.
//
import Foundation
import Vision

func recognizeText(_ path :String, completion: @escaping ([String]) -> Void) async {
	let handler = VNImageRequestHandler(url: URL(fileURLWithPath: path, isDirectory: false))
	let request = VNRecognizeTextRequest { request, error in
		let result = request.results as! [VNRecognizedTextObservation]
		let results = result.compactMap { $0.topCandidates(1).first?.string }
		DispatchQueue.global().async { completion(results) }
	}

	request.recognitionLevel = .accurate
	request.recognitionLanguages = ["zh-CN"]
	
	do {
		try handler.perform([request])
	} catch {
		fatalError(error.localizedDescription)
	}
}

let path = CommandLine.arguments[1]
await recognizeText(path) { $0.forEach { print($0) } }
