(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["default-footer"],{"3a66":function(t,e,i){"use strict";i.d(e,"a",(function(){return a}));var o=i("fe6c"),n=i("58df");function a(t){var e=arguments.length>1&&void 0!==arguments[1]?arguments[1]:[];return Object(n["a"])(Object(o["b"])(["absolute","fixed"])).extend({name:"applicationable",props:{app:Boolean},computed:{applicationProperty:function(){return t}},watch:{app:function(t,e){e?this.removeApplication(!0):this.callUpdate()},applicationProperty:function(t,e){this.$vuetify.application.unregister(this._uid,e)}},activated:function(){this.callUpdate()},created:function(){for(var t=0,i=e.length;t<i;t++)this.$watch(e[t],this.callUpdate);this.callUpdate()},mounted:function(){this.callUpdate()},deactivated:function(){this.removeApplication()},destroyed:function(){this.removeApplication()},methods:{callUpdate:function(){this.app&&this.$vuetify.application.register(this._uid,this.applicationProperty,this.updateApplication())},removeApplication:function(){var t=arguments.length>0&&void 0!==arguments[0]&&arguments[0];(t||this.app)&&this.$vuetify.application.unregister(this._uid,this.applicationProperty)},updateApplication:function(){return 0}}})}},"51ef":function(t,e,i){"use strict";i.r(e);var o=function(){var t=this,e=t.$createElement,i=t._self._c||e;return i("v-footer",{attrs:{id:"default-footer",color:"transparent",absolute:"",app:"",inset:""}})},n=[],a={name:"DefaultFooter"},s=a,r=i("2877"),c=i("6544"),p=i.n(c),u=i("5530"),l=(i("a9e3"),i("c7cd"),i("b5b6"),i("8dd9")),h=i("3a66"),d=i("d10f"),f=i("58df"),b=i("80d2"),v=Object(f["a"])(l["a"],Object(h["a"])("footer",["height","inset"]),d["a"]).extend({name:"v-footer",props:{height:{default:"auto",type:[Number,String]},inset:Boolean,padless:Boolean,tag:{type:String,default:"footer"}},computed:{applicationProperty:function(){return this.inset?"insetFooter":"footer"},classes:function(){return Object(u["a"])(Object(u["a"])({},l["a"].options.computed.classes.call(this)),{},{"v-footer--absolute":this.absolute,"v-footer--fixed":!this.absolute&&(this.app||this.fixed),"v-footer--padless":this.padless,"v-footer--inset":this.inset})},computedBottom:function(){if(this.isPositioned)return this.app?this.$vuetify.application.bottom:0},computedLeft:function(){if(this.isPositioned)return this.app&&this.inset?this.$vuetify.application.left:0},computedRight:function(){if(this.isPositioned)return this.app&&this.inset?this.$vuetify.application.right:0},isPositioned:function(){return Boolean(this.absolute||this.fixed||this.app)},styles:function(){var t=parseInt(this.height);return Object(u["a"])(Object(u["a"])({},l["a"].options.computed.styles.call(this)),{},{height:isNaN(t)?t:Object(b["h"])(t),left:Object(b["h"])(this.computedLeft),right:Object(b["h"])(this.computedRight),bottom:Object(b["h"])(this.computedBottom)})}},methods:{updateApplication:function(){var t=parseInt(this.height);return isNaN(t)?this.$el?this.$el.clientHeight:0:t}},render:function(t){var e=this.setBackgroundColor(this.color,{staticClass:"v-footer",class:this.classes,style:this.styles});return t(this.tag,e,this.$slots.default)}}),m=Object(r["a"])(s,o,n,!1,null,null,null);e["default"]=m.exports;p()(m,{VFooter:v})},b5b6:function(t,e,i){},d10f:function(t,e,i){"use strict";var o=i("2b0e");e["a"]=o["a"].extend({name:"ssr-bootable",data:function(){return{isBooted:!1}},mounted:function(){var t=this;window.requestAnimationFrame((function(){t.$el.setAttribute("data-booted","true"),t.isBooted=!0}))}})}}]);
//# sourceMappingURL=default-footer.d700febd.js.map