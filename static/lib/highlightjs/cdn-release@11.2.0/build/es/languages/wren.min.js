var hljsGrammar=(()=>{"use strict";function e(e){
return e?"string"==typeof e?e:e.source:null}function a(...a){
return a.map((a=>e(a))).join("")}function s(...a){return"("+((e=>{
const a=e[e.length-1]
;return"object"==typeof a&&a.constructor===Object?(e.splice(e.length-1,1),a):{}
})(a).capture?"":"?:")+a.map((a=>e(a))).join("|")+")"}return e=>{
const t=/[a-zA-Z]\w*/,n=["as","break","class","construct","continue","else","for","foreign","if","import","in","is","return","static","var","while"],c=["true","false","null"],r=["this","super"],i=["-","~",/\*/,"%",/\.\.\./,/\.\./,/\+/,"<<",">>",">=","<=","<",">",/\^/,/!=/,/!/,/\bis\b/,"==","&&","&",/\|\|/,/\|/,/\?:/,"="],o={
relevance:0,match:a(/\b(?!(if|while|for|else|super)\b)/,t,/(?=\s*[({])/),
className:"title.function"},l={
match:a(s(a(/\b(?!(if|while|for|else|super)\b)/,t),s(...i)),/(?=\s*\([^)]+\)\s*\{)/),
className:"title.function",starts:{contains:[{begin:/\(/,end:/\)/,contains:[{
relevance:0,scope:"params",match:t}]}]}},m={variants:[{
match:[/class\s+/,t,/\s+is\s+/,t]},{match:[/class\s+/,t]}],scope:{
2:"title.class",4:"title.class.inherited"},keywords:n},u={relevance:0,
match:s(...i),className:"operator"},p={className:"property",
begin:a(/\./,(b=t,a("(?=",b,")"))),end:t,excludeBegin:!0,relevance:0};var b
;const h={relevance:0,match:a(/\b_/,t),scope:"variable"},g={relevance:0,
match:/\b[A-Z]+[a-z]+([A-Z]+[a-z]+)*/,scope:"title.class",keywords:{
_:["Bool","Class","Fiber","Fn","List","Map","Null","Num","Object","Range","Sequence","String","System"]
}},f=e.C_NUMBER_MODE,v={match:[t,/\s*/,/=/,/\s*/,/\(/,t,/\)\s*\{/],scope:{
1:"title.function",3:"operator",6:"params"}},d=e.COMMENT(/\/\*\*/,/\*\//,{
contains:[{match:/@[a-z]+/,scope:"doctag"},"self"]}),N={scope:"subst",
begin:/%\(/,end:/\)/,contains:[f,g,o,h,u]},_={scope:"string",begin:/"/,end:/"/,
contains:[N,{scope:"char.escape",variants:[{match:/\\\\|\\["0%abefnrtv]/},{
match:/\\x[0-9A-F]{2}/},{match:/\\u[0-9A-F]{4}/},{match:/\\U[0-9A-F]{8}/}]}]}
;N.contains.push(_);const M={relevance:0,
match:a("\\b(?!",[...n,...r,...c].join("|"),"\\b)",/[a-zA-Z_]\w*(?:[?!]|\b)/),
className:"variable"};return{name:"Wren",keywords:{keyword:n,
"variable.language":r,literal:c},contains:[{scope:"comment",variants:[{
begin:[/#!?/,/[A-Za-z_]+(?=\()/],beginScope:{},keywords:{literal:c},contains:[],
end:/\)/},{begin:[/#!?/,/[A-Za-z_]+/],beginScope:{},end:/$/}]},f,_,{
className:"string",begin:/"""/,end:/"""/
},d,e.C_LINE_COMMENT_MODE,e.C_BLOCK_COMMENT_MODE,g,m,v,l,o,u,h,p,M]}}})()
;export default hljsGrammar;