import { readInput, writeOutput } from "./common.js";
import katex from "./katex.min.js";

const args = readInput();

// https://katex.org/docs/options
const options = {
	displayMode: args.displayMode ?? false,
	output: 'html',
	throwOnError: false,
};

const tex = args.tex ?? '';

writeOutput(katex.renderToString(tex, options));
