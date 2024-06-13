import * as std from "std";

const stdin = std.in.readAsString();
const args = JSON.parse(stdin);

// https://katex.org/docs/options
const options = {
	displayMode: args.displayMode ?? false,
	output: 'html',
	throwOnError: false,
};

const tex = args.tex ?? '';

std.out.puts(katex.renderToString(tex, options));
