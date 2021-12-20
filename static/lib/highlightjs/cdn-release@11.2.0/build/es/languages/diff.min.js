var hljsGrammar=(()=>{"use strict";function e(...e){return"("+((e=>{
const a=e[e.length-1]
;return"object"==typeof a&&a.constructor===Object?(e.splice(e.length-1,1),a):{}
})(e).capture?"":"?:")+e.map((e=>{return(a=e)?"string"==typeof a?a:a.source:null
;var a})).join("|")+")"}return a=>({name:"Diff",aliases:["patch"],contains:[{
className:"meta",relevance:10,
match:e(/^@@ +-\d+,\d+ +\+\d+,\d+ +@@/,/^\*\*\* +\d+,\d+ +\*\*\*\*$/,/^--- +\d+,\d+ +----$/)
},{className:"comment",variants:[{
begin:e(/Index: /,/^index/,/={3,}/,/^-{3}/,/^\*{3} /,/^\+{3}/,/^diff --git/),
end:/$/},{match:/^\*{15}$/}]},{className:"addition",begin:/^\+/,end:/$/},{
className:"deletion",begin:/^-/,end:/$/},{className:"addition",begin:/^!/,
end:/$/}]})})();export default hljsGrammar;