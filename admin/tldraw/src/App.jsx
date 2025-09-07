import { useCallback, useEffect, useState } from 'react'
import { Tldraw, getSnapshot, useEditor } from 'tldraw'
import 'tldraw/tldraw.css'
import './App.css'

function SnapshotToolbar() {
	const editor = useEditor();

	const save = useCallback(async () => {
		const shapeIds = editor.getCurrentPageShapeIds();
		if(shapeIds.size <= 0) {
			alert('不能保存，因为画布上没有任何内容。');
			return false;
		} 

		const { blob: lightBlob, width, height } = await editor.toImage([...shapeIds], {
			background: false,
			format: 'svg',
			darkMode: false,
			padding: 16,
		});
		const { blob: darkBlob } = await editor.toImage([...shapeIds], {
			background: false,
			format: 'svg',
			darkMode: true,
			padding: 16,
		});

		console.log('image length:', lightBlob.size, width, height);
		// console.log((await blob.bytes()).toBase64());

		const { document, session } = getSnapshot(editor.store);

		/** @type {(state: string, lightBlob: Blob, darkBlob: Blob) => Promise<boolean>} */
		const saver = editor.__saveSnapshot;
		return await saver(
			JSON.stringify({ document, session }),
			lightBlob, darkBlob,
		);
	}, [editor])

	const [showCheckMark, setShowCheckMark] = useState(false)
	useEffect(() => {
		if (showCheckMark) {
			const timeout = setTimeout(() => {
				setShowCheckMark(false)
			}, 1000)
			return () => clearTimeout(timeout)
		}
		return
	})

	return (
		<div className='share-buttons'>
			<span
				style={{
					display: 'inline-block',
					transition: 'transform 0.2s ease, opacity 0.2s ease',
					transform: showCheckMark ? `scale(1)` : `scale(0.5)`,
					opacity: showCheckMark ? 1 : 0,
				}}
			>✅ 已保存</span>
			<button style={{backgroundColor: 'var(--color-selected', color: 'white'}}
				onClick={async () => {
					if(await save()) {
						setShowCheckMark(true)
					}
				}}
			>保存</button>
		</div>
	)
}

export default function App({snapshotJSON, saveSnapshot}) {
	return (
		<div className='tldraw-app'>
			<Tldraw
				snapshot={JSON.parse(snapshotJSON ?? '{}')}
				onMount={ed => {
					ed.__saveSnapshot = saveSnapshot;
				}}
				options={{
					maxPages: 1,
				}}
				components={{
					SharePanel: SnapshotToolbar,
				}}
			/>
		</div>
	)
}
