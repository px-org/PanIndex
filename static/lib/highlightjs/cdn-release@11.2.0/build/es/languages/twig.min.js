var hljsGrammar=(()=>{"use strict";return e=>{
var a="attribute block constant cycle date dump include max min parent random range source template_from_string",t={
beginKeywords:a,keywords:{name:a},relevance:0,contains:[{className:"params",
begin:"\\(",end:"\\)"}]},n={begin:/\|[A-Za-z_]+:?/,
keywords:"abs batch capitalize column convert_encoding date date_modify default escape filter first format inky_to_html inline_css join json_encode keys last length lower map markdown merge nl2br number_format raw reduce replace reverse round slice sort spaceless split striptags title trim upper url_encode",
contains:[t]
},s="apply autoescape block deprecated do embed extends filter flush for from if import include macro sandbox set use verbatim with"
;return s=s+" "+s.split(" ").map((e=>"end"+e)).join(" "),{name:"Twig",
aliases:["craftcms"],case_insensitive:!0,subLanguage:"xml",
contains:[e.COMMENT(/\{#/,/#\}/),{className:"template-tag",begin:/\{%/,
end:/%\}/,contains:[{className:"name",begin:/\w+/,keywords:s,starts:{
endsWithParent:!0,contains:[n,t],relevance:0}}]},{className:"template-variable",
begin:/\{\{/,end:/\}\}/,contains:["self",n,t]}]}}})()
;export default hljsGrammar;