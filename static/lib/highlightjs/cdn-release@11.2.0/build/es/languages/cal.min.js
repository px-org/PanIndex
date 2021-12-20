var hljsGrammar=(()=>{"use strict";return e=>{
const n="div mod in and or not xor asserterror begin case do downto else end exit for if of repeat then to until while with var",r=[e.C_LINE_COMMENT_MODE,e.COMMENT(/\{/,/\}/,{
relevance:0}),e.COMMENT(/\(\*/,/\*\)/,{relevance:10})],a={className:"string",
begin:/'/,end:/'/,contains:[{begin:/''/}]},s={className:"string",begin:/(#\d+)+/
},t={className:"function",beginKeywords:"procedure",end:/[:;]/,
keywords:"procedure|10",contains:[e.TITLE_MODE,{className:"params",begin:/\(/,
end:/\)/,keywords:n,contains:[a,s]}].concat(r)},i={className:"class",
begin:"OBJECT (Table|Form|Report|Dataport|Codeunit|XMLport|MenuSuite|Page|Query) (\\d+) ([^\\r\\n]+)",
returnBegin:!0,contains:[e.TITLE_MODE,t]};return{name:"C/AL",
case_insensitive:!0,keywords:{keyword:n,literal:"false true"},illegal:/\/\*/,
contains:[a,s,{className:"number",begin:"\\b\\d+(\\.\\d+)?(DT|D|T)",relevance:0
},{className:"string",begin:'"',end:'"'},e.NUMBER_MODE,i,t]}}})()
;export default hljsGrammar;