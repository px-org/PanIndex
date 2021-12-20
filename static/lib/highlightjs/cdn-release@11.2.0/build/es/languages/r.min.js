var hljsGrammar=(()=>{"use strict";function e(e){
return e?"string"==typeof e?e:e.source:null}function a(...a){return"("+((e=>{
const a=e[e.length-1]
;return"object"==typeof a&&a.constructor===Object?(e.splice(e.length-1,1),a):{}
})(a).capture?"":"?:")+a.map((a=>e(a))).join("|")+")"}return n=>{
const s=/(?:(?:[a-zA-Z]|\.[._a-zA-Z])[._a-zA-Z0-9]*)|\.(?!\d)/,i=a(/0[xX][0-9a-fA-F]+\.[0-9a-fA-F]*[pP][+-]?\d+i?/,/0[xX][0-9a-fA-F]+(?:[pP][+-]?\d+)?[Li]?/,/(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?[Li]?/),t=/[=!<>:]=|\|\||&&|:::?|<-|<<-|->>|->|\|>|[-+*\/?!$&|:<=>@^~]|\*\*/,r=a(/[()]/,/[{}]/,/\[\[/,/[[\]]/,/\\/,/,/)
;return{name:"R",keywords:{$pattern:s,
keyword:"function if in break next repeat else for while",
literal:"NULL NA TRUE FALSE Inf NaN NA_integer_|10 NA_real_|10 NA_character_|10 NA_complex_|10",
built_in:"LETTERS letters month.abb month.name pi T F abs acos acosh all any anyNA Arg as.call as.character as.complex as.double as.environment as.integer as.logical as.null.default as.numeric as.raw asin asinh atan atanh attr attributes baseenv browser c call ceiling class Conj cos cosh cospi cummax cummin cumprod cumsum digamma dim dimnames emptyenv exp expression floor forceAndCall gamma gc.time globalenv Im interactive invisible is.array is.atomic is.call is.character is.complex is.double is.environment is.expression is.finite is.function is.infinite is.integer is.language is.list is.logical is.matrix is.na is.name is.nan is.null is.numeric is.object is.pairlist is.raw is.recursive is.single is.symbol lazyLoadDBfetch length lgamma list log max min missing Mod names nargs nzchar oldClass on.exit pos.to.env proc.time prod quote range Re rep retracemem return round seq_along seq_len seq.int sign signif sin sinh sinpi sqrt standardGeneric substitute sum switch tan tanh tanpi tracemem trigamma trunc unclass untracemem UseMethod xtfrm"
},contains:[n.COMMENT(/#'/,/$/,{contains:[{scope:"doctag",match:/@examples/,
starts:{
end:(c=a(/\n^#'\s*(?=@[a-zA-Z]+)/,/\n^(?!#')/),((...a)=>a.map((a=>e(a))).join(""))("(?=",c,")")),
endsParent:!0}},{scope:"doctag",begin:"@param",end:/$/,contains:[{
scope:"variable",variants:[{match:s},{match:/`(?:\\.|[^`\\])+`/}],endsParent:!0
}]},{scope:"doctag",match:/@[a-zA-Z]+/},{scope:"keyword",match:/\\[a-zA-Z]+/}]
}),n.HASH_COMMENT_MODE,{scope:"string",contains:[n.BACKSLASH_ESCAPE],
variants:[n.END_SAME_AS_BEGIN({begin:/[rR]"(-*)\(/,end:/\)(-*)"/
}),n.END_SAME_AS_BEGIN({begin:/[rR]"(-*)\{/,end:/\}(-*)"/
}),n.END_SAME_AS_BEGIN({begin:/[rR]"(-*)\[/,end:/\](-*)"/
}),n.END_SAME_AS_BEGIN({begin:/[rR]'(-*)\(/,end:/\)(-*)'/
}),n.END_SAME_AS_BEGIN({begin:/[rR]'(-*)\{/,end:/\}(-*)'/
}),n.END_SAME_AS_BEGIN({begin:/[rR]'(-*)\[/,end:/\](-*)'/}),{begin:'"',end:'"',
relevance:0},{begin:"'",end:"'",relevance:0}]},{relevance:0,variants:[{scope:{
1:"operator",2:"number"},match:[t,i]},{scope:{1:"operator",2:"number"},
match:[/%[^%]*%/,i]},{scope:{1:"punctuation",2:"number"},match:[r,i]},{scope:{
2:"number"},match:[/[^a-zA-Z0-9._]|^/,i]}]},{scope:{3:"operator"},
match:[s,/\s+/,/<-/,/\s+/]},{scope:"operator",relevance:0,variants:[{match:t},{
match:/%[^%]*%/}]},{scope:"punctuation",relevance:0,match:r},{begin:"`",end:"`",
contains:[{begin:/\\./}]}]};var c}})();export default hljsGrammar;