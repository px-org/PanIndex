;
(function(g, fn) {
	var version = "1.4.00",
		pdfjsVersion = "2.3.200";
	console.log("pdfh5.js v" + version + "  https://www.gjtool.cn")
	if (typeof require !== 'undefined') {
		if (g.$ === undefined) {
			g.$ = require('./jquery-2.1.1.min.js');
		}
		g.pdfjsWorker = require('./pdf.worker.js');
		g.pdfjsLib = require('./pdf.js');
	}
	var pdfjsLib = g.pdfjsLib,
		$ = g.$,
		pdfjsWorker = g.pdfjsWorker;
	if (typeof define === 'function' && define.amd) {
		define(function() {
			return fn(g, pdfjsWorker, pdfjsLib, $, version)
		})
	} else if (typeof module !== 'undefined' && module.exports) {
		module.exports = fn(g, pdfjsWorker, pdfjsLib, $, version)
	} else {
		g.Pdfh5 = fn(g, pdfjsWorker, pdfjsLib, $, version)
	}
})(typeof window !== 'undefined' ? window : this, function(g, pdfjsWorker, pdfjsLib, $, version) {
	'use strict';

	var _createClass = function() {
		function defineProperties(target, props) {
			for (var i = 0; i < props.length; i++) {
				var descriptor = props[i];
				descriptor.enumerable = descriptor.enumerable || false;
				descriptor.configurable = true;
				if ("value" in descriptor) descriptor.writable = true;
				Object.defineProperty(target, descriptor.key, descriptor);
			}
		}
		return function(Constructor, protoProps, staticProps) {
			if (protoProps) defineProperties(Constructor.prototype, protoProps);
			if (staticProps) defineProperties(Constructor, staticProps);
			return Constructor;
		};
	}();

	function _classCallCheck(instance, Constructor) {
		if (!(instance instanceof Constructor)) {
			throw new TypeError("Cannot call a class as a function");
		}
	}

	var renderTextLayer = pdfjsLib.renderTextLayer;
	var EXPAND_DIVS_TIMEOUT = 300; // ms

	var TextLayerBuilder = function() {
		function TextLayerBuilder(_ref) {
			var textLayerDiv = _ref.textLayerDiv;
			var eventBus = _ref.eventBus;
			var pageIndex = _ref.pageIndex;
			var viewport = _ref.viewport;
			var _ref$findController = _ref.findController;
			var findController = _ref$findController === undefined ? null : _ref$findController;
			var _ref$enhanceTextSelec = _ref.enhanceTextSelection;
			var enhanceTextSelection = _ref$enhanceTextSelec === undefined ? false : _ref$enhanceTextSelec;

			_classCallCheck(this, TextLayerBuilder);

			this.textLayerDiv = textLayerDiv;
			this.eventBus = eventBus;
			this.textContent = null;
			this.textContentItemsStr = [];
			this.textContentStream = null;
			this.renderingDone = false;
			this.pageIdx = pageIndex;
			this.pageNumber = this.pageIdx + 1;
			this.matches = [];
			this.viewport = viewport;
			this.textDivs = [];
			this.findController = findController;
			this.textLayerRenderTask = null;
			this.enhanceTextSelection = enhanceTextSelection;

			this._onUpdateTextLayerMatches = null;
			this._bindMouse();
		}

		/**
		 * @private
		 */


		_createClass(TextLayerBuilder, [{
			key: "_finishRendering",
			value: function _finishRendering() {
				this.renderingDone = true;

				if (!this.enhanceTextSelection) {
					var endOfContent = document.createElement("div");
					endOfContent.className = "endOfContent";
					this.textLayerDiv.appendChild(endOfContent);
				}
				if (this.eventBus) {
					this.eventBus.dispatch("textlayerrendered", {
						source: this,
						pageNumber: this.pageNumber,
						numTextDivs: this.textDivs.length
					});
				}
			}

			/**
			 * Renders the text layer.
			 *
			 * @param {number} [timeout] - Wait for a specified amount of milliseconds
			 *                             before rendering.
			 */

		}, {
			key: "render",
			value: function render() {
				var _this = this;

				var timeout = arguments.length <= 0 || arguments[0] === undefined ? 0 :
					arguments[0];

				if (!(this.textContent || this.textContentStream) || this.renderingDone) {
					return;
				}
				this.cancel();

				this.textDivs = [];
				var textLayerFrag = document.createDocumentFragment();
				this.textLayerRenderTask = renderTextLayer({
					textContent: this.textContent,
					textContentStream: this.textContentStream,
					container: textLayerFrag,
					viewport: this.viewport,
					textDivs: this.textDivs,
					textContentItemsStr: this.textContentItemsStr,
					timeout: timeout,
					enhanceTextSelection: this.enhanceTextSelection
				});
				this.textLayerRenderTask.promise.then(function() {
					_this.textLayerDiv.appendChild(textLayerFrag);
					_this._finishRendering();
					_this._updateMatches();
				}, function(reason) {
					// Cancelled or failed to render text layer; skipping errors.
				});

				if (!this._onUpdateTextLayerMatches) {
					this._onUpdateTextLayerMatches = function(evt) {
						if (evt.pageIndex === _this.pageIdx || evt.pageIndex === -1) {
							_this._updateMatches();
						}
					};
					if (this.eventBus) {
						this.eventBus._on("updatetextlayermatches", this
							._onUpdateTextLayerMatches);
					}
				}
			}

			/**
			 * Cancel rendering of the text layer.
			 */

		}, {
			key: "cancel",
			value: function cancel() {
				if (this.textLayerRenderTask) {
					this.textLayerRenderTask.cancel();
					this.textLayerRenderTask = null;
				}
				if (this._onUpdateTextLayerMatches) {
					this.eventBus._off("updatetextlayermatches", this
						._onUpdateTextLayerMatches);
					this._onUpdateTextLayerMatches = null;
				}
			}
		}, {
			key: "setTextContentStream",
			value: function setTextContentStream(readableStream) {
				this.cancel();
				this.textContentStream = readableStream;
			}
		}, {
			key: "setTextContent",
			value: function setTextContent(textContent) {
				this.cancel();
				this.textContent = textContent;
			}
		}, {
			key: "_convertMatches",
			value: function _convertMatches(matches, matchesLength) {
				// Early exit if there is nothing to convert.
				if (!matches) {
					return [];
				}
				var findController = this.findController;
				var textContentItemsStr = this.textContentItemsStr;


				var i = 0,
					iIndex = 0;
				var end = textContentItemsStr.length - 1;
				var queryLen = findController.state.query.length;
				var result = [];

				for (var m = 0, mm = matches.length; m < mm; m++) {
					// Calculate the start position.
					var matchIdx = matches[m];

					// Loop over the divIdxs.
					while (i !== end && matchIdx >= iIndex + textContentItemsStr[i].length) {
						iIndex += textContentItemsStr[i].length;
						i++;
					}

					if (i === textContentItemsStr.length) {
						console.error("Could not find a matching mapping");
					}

					var match = {
						begin: {
							divIdx: i,
							offset: matchIdx - iIndex
						}
					};

					// Calculate the end position.
					if (matchesLength) {
						// Multiterm search.
						matchIdx += matchesLength[m];
					} else {
						// Phrase search.
						matchIdx += queryLen;
					}

					// Somewhat the same array as above, but use > instead of >= to get
					// the end position right.
					while (i !== end && matchIdx > iIndex + textContentItemsStr[i].length) {
						iIndex += textContentItemsStr[i].length;
						i++;
					}

					match.end = {
						divIdx: i,
						offset: matchIdx - iIndex
					};
					result.push(match);
				}
				return result;
			}
		}, {
			key: "_renderMatches",
			value: function _renderMatches(matches) {
				// Early exit if there is nothing to render.
				if (matches.length === 0) {
					return;
				}
				var findController = this.findController;
				var pageIdx = this.pageIdx;
				var textContentItemsStr = this.textContentItemsStr;
				var textDivs = this.textDivs;


				var isSelectedPage = pageIdx === findController.selected.pageIdx;
				var selectedMatchIdx = findController.selected.matchIdx;
				var highlightAll = findController.state.highlightAll;
				var prevEnd = null;
				var infinity = {
					divIdx: -1,
					offset: undefined
				};

				function beginText(begin, className) {
					var divIdx = begin.divIdx;
					textDivs[divIdx].textContent = "";
					appendTextToDiv(divIdx, 0, begin.offset, className);
				}

				function appendTextToDiv(divIdx, fromOffset, toOffset, className) {
					var div = textDivs[divIdx];
					var content = textContentItemsStr[divIdx].substring(fromOffset, toOffset);
					var node = document.createTextNode(content);
					if (className) {
						var span = document.createElement("span");
						span.className = className;
						span.appendChild(node);
						div.appendChild(span);
						return;
					}
					div.appendChild(node);
				}

				var i0 = selectedMatchIdx,
					i1 = i0 + 1;
				if (highlightAll) {
					i0 = 0;
					i1 = matches.length;
				} else if (!isSelectedPage) {
					// Not highlighting all and this isn't the selected page, so do nothing.
					return;
				}

				for (var i = i0; i < i1; i++) {
					var match = matches[i];
					var begin = match.begin;
					var end = match.end;
					var isSelected = isSelectedPage && i === selectedMatchIdx;
					var highlightSuffix = isSelected ? " selected" : "";

					if (isSelected) {
						// Attempt to scroll the selected match into view.
						findController.scrollMatchIntoView({
							element: textDivs[begin.divIdx],
							pageIndex: pageIdx,
							matchIndex: selectedMatchIdx
						});
					}

					// Match inside new div.
					if (!prevEnd || begin.divIdx !== prevEnd.divIdx) {
						// If there was a previous div, then add the text at the end.
						if (prevEnd !== null) {
							appendTextToDiv(prevEnd.divIdx, prevEnd.offset, infinity.offset);
						}
						// Clear the divs and set the content until the starting point.
						beginText(begin);
					} else {
						appendTextToDiv(prevEnd.divIdx, prevEnd.offset, begin.offset);
					}

					if (begin.divIdx === end.divIdx) {
						appendTextToDiv(begin.divIdx, begin.offset, end.offset, "highlight" +
							highlightSuffix);
					} else {
						appendTextToDiv(begin.divIdx, begin.offset, infinity.offset,
							"highlight begin" + highlightSuffix);
						for (var n0 = begin.divIdx + 1, n1 = end.divIdx; n0 < n1; n0++) {
							textDivs[n0].className = "highlight middle" + highlightSuffix;
						}
						beginText(end, "highlight end" + highlightSuffix);
					}
					prevEnd = end;
				}

				if (prevEnd) {
					appendTextToDiv(prevEnd.divIdx, prevEnd.offset, infinity.offset);
				}
			}
		}, {
			key: "_updateMatches",
			value: function _updateMatches() {
				// Only show matches when all rendering is done.
				if (!this.renderingDone) {
					return;
				}
				var findController = this.findController;
				var matches = this.matches;
				var pageIdx = this.pageIdx;
				var textContentItemsStr = this.textContentItemsStr;
				var textDivs = this.textDivs;

				var clearedUntilDivIdx = -1;

				// Clear all current matches.
				for (var i = 0, ii = matches.length; i < ii; i++) {
					var match = matches[i];
					var begin = Math.max(clearedUntilDivIdx, match.begin.divIdx);
					for (var n = begin, end = match.end.divIdx; n <= end; n++) {
						var div = textDivs[n];
						div.textContent = textContentItemsStr[n];
						div.className = "";
					}
					clearedUntilDivIdx = match.end.divIdx + 1;
				}

				if (!findController || !findController.highlightMatches) {
					return;
				}
				// Convert the matches on the `findController` into the match format
				// used for the textLayer.
				var pageMatches = findController.pageMatches[pageIdx] || null;
				var pageMatchesLength = findController.pageMatchesLength[pageIdx] || null;

				this.matches = this._convertMatches(pageMatches, pageMatchesLength);
				this._renderMatches(this.matches);
			}

			/**
			 * Improves text selection by adding an additional div where the mouse was
			 * clicked. This reduces flickering of the content if the mouse is slowly
			 * dragged up or down.
			 *
			 * @private
			 */

		}, {
			key: "_bindMouse",
			value: function _bindMouse() {
				var _this2 = this;

				var div = this.textLayerDiv;
				var expandDivsTimer = null;

				div.addEventListener("mousedown", function(evt) {
					if (_this2.enhanceTextSelection && _this2.textLayerRenderTask) {
						_this2.textLayerRenderTask.expandTextDivs(true);
						if ((typeof PDFJSDev === "undefined" || !PDFJSDev.test(
								"MOZCENTRAL")) && expandDivsTimer) {
							clearTimeout(expandDivsTimer);
							expandDivsTimer = null;
						}
						return;
					}

					var end = div.querySelector(".endOfContent");
					if (!end) {
						return;
					}
					if (typeof PDFJSDev === "undefined" || !PDFJSDev.test(
							"MOZCENTRAL")) {
						// On non-Firefox browsers, the selection will feel better if the height
						// of the `endOfContent` div is adjusted to start at mouse click
						// location. This avoids flickering when the selection moves up.
						// However it does not work when selection is started on empty space.
						var adjustTop = evt.target !== div;
						if (typeof PDFJSDev === "undefined" || PDFJSDev.test(
								"GENERIC")) {
							adjustTop = adjustTop && window.getComputedStyle(end)
								.getPropertyValue("-moz-user-select") !== "none";
						}
						if (adjustTop) {
							var divBounds = div.getBoundingClientRect();
							var r = Math.max(0, (evt.pageY - divBounds.top) / divBounds
								.height);
							end.style.top = (r * 100).toFixed(2) + "%";
						}
					}
					end.classList.add("active");
				});

				div.addEventListener("mouseup", function() {
					if (_this2.enhanceTextSelection && _this2.textLayerRenderTask) {
						if (typeof PDFJSDev === "undefined" || !PDFJSDev.test(
								"MOZCENTRAL")) {
							expandDivsTimer = setTimeout(function() {
								if (_this2.textLayerRenderTask) {
									_this2.textLayerRenderTask.expandTextDivs(
										false);
								}
								expandDivsTimer = null;
							}, EXPAND_DIVS_TIMEOUT);
						} else {
							_this2.textLayerRenderTask.expandTextDivs(false);
						}
						return;
					}

					var end = div.querySelector(".endOfContent");
					if (!end) {
						return;
					}
					if (typeof PDFJSDev === "undefined" || !PDFJSDev.test(
							"MOZCENTRAL")) {
						end.style.top = "";
					}
					end.classList.remove("active");
				});
			}
		}]);

		return TextLayerBuilder;
	}();

	/**
	 * @implements IPDFTextLayerFactory
	 */


	var DefaultTextLayerFactory = function() {
		function DefaultTextLayerFactory() {
			_classCallCheck(this, DefaultTextLayerFactory);
		}

		_createClass(DefaultTextLayerFactory, [{
			key: "createTextLayerBuilder",

			/**
			 * @param {HTMLDivElement} textLayerDiv
			 * @param {number} pageIndex
			 * @param {PageViewport} viewport
			 * @param {boolean} enhanceTextSelection
			 * @param {EventBus} eventBus
			 * @returns {TextLayerBuilder}
			 */
			value: function createTextLayerBuilder(textLayerDiv, pageIndex, viewport) {
				var enhanceTextSelection = arguments.length <= 3 || arguments[3] === undefined ?
					false : arguments[3];
				var eventBus = arguments[4];

				return new TextLayerBuilder({
					textLayerDiv: textLayerDiv,
					pageIndex: pageIndex,
					viewport: viewport,
					enhanceTextSelection: enhanceTextSelection,
					eventBus: eventBus
				});
			}
		}]);

		return DefaultTextLayerFactory;
	}();

	g.TextLayerBuilder = TextLayerBuilder;
	g.DefaultTextLayerFactory = DefaultTextLayerFactory;

	var definePinchZoom = function($) {
		var PinchZoom = function(el, options, viewerContainer) {
				this.el = $(el);
				this.viewerContainer = viewerContainer;
				this.zoomFactor = 1;
				this.lastScale = 1;
				this.offset = {
					x: 0,
					y: 0
				};
				this.options = $.extend({}, this.defaults, options);
				this.options.zoomOutFactor = isNaN(options.zoomOutFactor) ? 1.2 : options.zoomOutFactor;
				this.options.animationDuration = isNaN(options.animationDuration) ? 300 : options
					.animationDuration;
				this.options.maxZoom = isNaN(options.maxZoom) ? 3 : options.maxZoom;
				this.options.minZoom = isNaN(options.minZoom) ? 0.8 : options.minZoom;
				this.setupMarkup();
				this.bindEvents();
				this.update();
				this.enable();
				this.height = 0;
				this.load = false;
				this.direction = null;
				this.clientY = null;
				this.lastclientY = null;
			},
			sum = function(a, b) {
				return a + b;
			},
			isCloseTo = function(value, expected) {
				return value > expected - 0.01 && value < expected + 0.01;
			};

		PinchZoom.prototype = {
			defaults: {
				tapZoomFactor: 3,
				zoomOutFactor: 1.2,
				animationDuration: 300,
				maxZoom: 5,
				minZoom: 0.8,
				lockDragAxis: false,
				use2d: true,
				zoomStartEventName: 'pz_zoomstart',
				zoomEndEventName: 'pz_zoomend',
				dragStartEventName: 'pz_dragstart',
				dragEndEventName: 'pz_dragend',
				doubleTapEventName: 'pz_doubletap'
			},
			handleDragStart: function(event) {
				this.el.trigger(this.options.dragStartEventName);
				this.stopAnimation();
				this.lastDragPosition = false;
				this.hasInteraction = true;
				this.handleDrag(event);
			},
			handleDrag: function(event) {

				if (this.zoomFactor > 1.0) {
					var touch = this.getTouches(event)[0];
					this.drag(touch, this.lastDragPosition, event);
					this.offset = this.sanitizeOffset(this.offset);
					this.lastDragPosition = touch;
				}
			},

			handleDragEnd: function() {
				this.el.trigger(this.options.dragEndEventName);
				this.end();
			},
			handleZoomStart: function(event) {
				this.el.trigger(this.options.zoomStartEventName);
				this.stopAnimation();
				this.lastScale = 1;
				this.nthZoom = 0;
				this.lastZoomCenter = false;
				this.hasInteraction = true;
			},
			handleZoom: function(event, newScale) {
				var touchCenter = this.getTouchCenter(this.getTouches(event)),
					scale = newScale / this.lastScale;
				this.lastScale = newScale;
				this.nthZoom += 1;
				if (this.nthZoom > 3) {

					this.scale(scale, touchCenter);
					this.drag(touchCenter, this.lastZoomCenter);
				}
				this.lastZoomCenter = touchCenter;
			},

			handleZoomEnd: function() {
				this.el.trigger(this.options.zoomEndEventName);
				this.end();
			},
			handleDoubleTap: function(event) {
				var center = this.getTouches(event)[0],
					zoomFactor = this.zoomFactor > 1 ? 1 : this.options.tapZoomFactor,
					startZoomFactor = this.zoomFactor,
					updateProgress = (function(progress) {
						this.scaleTo(startZoomFactor + progress * (zoomFactor - startZoomFactor),
							center);
					}).bind(this);

				if (this.hasInteraction) {
					return;
				}
				if (startZoomFactor > zoomFactor) {
					center = this.getCurrentZoomCenter();
				}

				this.animate(this.options.animationDuration, updateProgress, this.swing);
				this.el.trigger(this.options.doubleTapEventName);
			},
			sanitizeOffset: function(offset) {
				var maxX = (this.zoomFactor - 1) * this.getContainerX(),
					maxY = (this.zoomFactor - 1) * this.getContainerY(),
					maxOffsetX = Math.max(maxX, 0),
					maxOffsetY = Math.max(maxY, 0),
					minOffsetX = Math.min(maxX, 0),
					minOffsetY = Math.min(maxY, 0);

				var x = Math.min(Math.max(offset.x, minOffsetX), maxOffsetX),
					y = Math.min(Math.max(offset.y, minOffsetY), maxOffsetY);


				return {
					x: x,
					y: y
				};
			},
			scaleTo: function(zoomFactor, center) {
				this.scale(zoomFactor / this.zoomFactor, center);
			},
			scale: function(scale, center) {
				scale = this.scaleZoomFactor(scale);
				this.addOffset({
					x: (scale - 1) * (center.x + this.offset.x),
					y: (scale - 1) * (center.y + this.offset.y)
				});
				this.done && this.done.call(this, this.getInitialZoomFactor() * this.zoomFactor)
			},
			scaleZoomFactor: function(scale) {
				var originalZoomFactor = this.zoomFactor;
				this.zoomFactor *= scale;
				this.zoomFactor = Math.min(this.options.maxZoom, Math.max(this.zoomFactor, this.options
					.minZoom));
				return this.zoomFactor / originalZoomFactor;
			},
			drag: function(center, lastCenter, event) {
				if (lastCenter) {
					if (this.options.lockDragAxis) {
						if (Math.abs(center.x - lastCenter.x) > Math.abs(center.y - lastCenter.y)) {
							this.addOffset({
								x: -(center.x - lastCenter.x),
								y: 0
							});
						} else {
							this.addOffset({
								y: -(center.y - lastCenter.y),
								x: 0
							});
						}
					} else {
						if (center.y - lastCenter.y < 0) {
							this.direction = "down";
						} else if (center.y - lastCenter.y > 10) {
							this.direction = "up";
						}
						this.addOffset({
							y: -(center.y - lastCenter.y),
							x: -(center.x - lastCenter.x)
						});
					}
				}
			},
			getTouchCenter: function(touches) {
				return this.getVectorAvg(touches);
			},
			getVectorAvg: function(vectors) {
				return {
					x: vectors.map(function(v) {
						return v.x;
					}).reduce(sum) / vectors.length,
					y: vectors.map(function(v) {
						return v.y;
					}).reduce(sum) / vectors.length
				};
			},
			addOffset: function(offset) {
				this.offset = {
					x: this.offset.x + offset.x,
					y: this.offset.y + offset.y
				};
			},

			sanitize: function() {
				if (this.zoomFactor < this.options.zoomOutFactor) {
					this.zoomOutAnimation();
				} else if (this.isInsaneOffset(this.offset)) {
					this.sanitizeOffsetAnimation();
				}
			},
			isInsaneOffset: function(offset) {
				var sanitizedOffset = this.sanitizeOffset(offset);
				return sanitizedOffset.x !== offset.x ||
					sanitizedOffset.y !== offset.y;
			},
			sanitizeOffsetAnimation: function() {
				var targetOffset = this.sanitizeOffset(this.offset),
					startOffset = {
						x: this.offset.x,
						y: this.offset.y
					},
					updateProgress = (function(progress) {
						this.offset.x = startOffset.x + progress * (targetOffset.x - startOffset.x);
						this.offset.y = startOffset.y + progress * (targetOffset.y - startOffset.y);
						this.update();
					}).bind(this);

				this.animate(
					this.options.animationDuration,
					updateProgress,
					this.swing
				);
			},
			zoomOutAnimation: function() {
				var startZoomFactor = this.zoomFactor,
					zoomFactor = 1,
					center = this.getCurrentZoomCenter(),
					updateProgress = (function(progress) {
						this.scaleTo(startZoomFactor + progress * (zoomFactor - startZoomFactor),
							center);
					}).bind(this);

				this.animate(
					this.options.animationDuration,
					updateProgress,
					this.swing
				);
			},
			updateAspectRatio: function() {
				this.setContainerY(this.getContainerX() / this.getAspectRatio());
			},
			getInitialZoomFactor: function() {
				if (this.container[0] && this.el[0]) {
					return this.container[0].offsetWidth / this.el[0].offsetWidth;
				} else {
					return 0
				}
			},
			getAspectRatio: function() {
				if (this.el[0]) {
					var offsetHeight = this.el[0].offsetHeight;
					return this.container[0].offsetWidth / offsetHeight;
				} else {
					return 0
				}

			},
			getCurrentZoomCenter: function() {
				var length = this.container[0].offsetWidth * this.zoomFactor,
					offsetLeft = this.offset.x,
					offsetRight = length - offsetLeft - this.container[0].offsetWidth,
					widthOffsetRatio = offsetLeft / offsetRight,
					centerX = widthOffsetRatio * this.container[0].offsetWidth / (widthOffsetRatio + 1),

					height = this.container[0].offsetHeight * this.zoomFactor,
					offsetTop = this.offset.y,
					offsetBottom = height - offsetTop - this.container[0].offsetHeight,
					heightOffsetRatio = offsetTop / offsetBottom,
					centerY = heightOffsetRatio * this.container[0].offsetHeight / (heightOffsetRatio +
						1);

				if (offsetRight === 0) {
					centerX = this.container[0].offsetWidth;
				}
				if (offsetBottom === 0) {
					centerY = this.container[0].offsetHeight;
				}

				return {
					x: centerX,
					y: centerY
				};
			},

			canDrag: function() {
				return !isCloseTo(this.zoomFactor, 1);
			},

			getTouches: function(event) {
				var position = this.container.offset();
				return Array.prototype.slice.call(event.touches).map(function(touch) {
					return {
						x: touch.pageX - position.left,
						y: touch.pageY - position.top
					};
				});
			},
			animate: function(duration, framefn, timefn, callback) {
				var startTime = new Date().getTime(),
					renderFrame = (function() {
						if (!this.inAnimation) {
							return;
						}
						var frameTime = new Date().getTime() - startTime,
							progress = frameTime / duration;
						if (frameTime >= duration) {
							framefn(1);
							if (callback) {
								callback();
							}
							this.update();
							this.stopAnimation();
						} else {
							if (timefn) {
								progress = timefn(progress);
							}
							framefn(progress);
							this.update();
							requestAnimationFrame(renderFrame);
						}
					}).bind(this);
				this.inAnimation = true;
				requestAnimationFrame(renderFrame);
			},
			stopAnimation: function() {
				this.inAnimation = false;

			},
			swing: function(p) {
				return -Math.cos(p * Math.PI) / 2 + 0.5;
			},

			getContainerX: function() {
				if (this.el[0]) {
					return this.el[0].offsetWidth;
				} else {
					return 0;
				}
			},

			getContainerY: function() {
				return this.el[0].offsetHeight;
			},
			setContainerY: function(y) {
				y = y.toFixed(2);
				return this.container.height(y);
			},
			setupMarkup: function() {
				this.container = $('<div class="pinch-zoom-container"></div>');
				this.el.before(this.container);
				this.container.append(this.el);

				this.container.css({
					'position': 'relative',
				});

				this.el.css({
					'-webkit-transform-origin': '0% 0%',
					'-moz-transform-origin': '0% 0%',
					'-ms-transform-origin': '0% 0%',
					'-o-transform-origin': '0% 0%',
					'transform-origin': '0% 0%',
					'position': 'relative'
				});

			},

			end: function() {
				this.hasInteraction = false;
				this.sanitize();
				this.update();

			},
			bindEvents: function() {
				detectGestures(this.container.eq(0), this, this.viewerContainer);
				$(g).on('resize', this.update.bind(this));
				$(this.el).find('img').on('load', this.update.bind(this));

			},
			update: function() {

				if (this.updatePlaned) {
					return;
				}
				this.updatePlaned = true;
				setTimeout((function() {
					this.updatePlaned = false;
					this.updateAspectRatio();
					var zoomFactor = this.getInitialZoomFactor() * this.zoomFactor,
						offsetX = (-this.offset.x / zoomFactor).toFixed(3),
						offsetY = (-this.offset.y / zoomFactor).toFixed(3);
					this.lastclientY = offsetY;

					var transform3d = 'scale3d(' + zoomFactor + ', ' + zoomFactor + ',1) ' +
						'translate3d(' + offsetX + 'px,' + offsetY + 'px,0px)',
						transform2d = 'scale(' + zoomFactor + ', ' + zoomFactor + ') ' +
						'translate(' + offsetX + 'px,' + offsetY + 'px)',
						removeClone = (function() {
							if (this.clone) {
								this.clone.remove();
								delete this.clone;
							}
						}).bind(this);
					if (!this.options.use2d || this.hasInteraction || this.inAnimation) {
						this.is3d = true;
						this.el.css({
							'-webkit-transform': transform3d,
							'-o-transform': transform2d,
							'-ms-transform': transform2d,
							'-moz-transform': transform2d,
							'transform': transform3d
						});
					} else {
						this.el.css({
							'-webkit-transform': transform2d,
							'-o-transform': transform2d,
							'-ms-transform': transform2d,
							'-moz-transform': transform2d,
							'transform': transform2d
						});
						this.is3d = false;
					}
				}).bind(this), 0);
			},
			enable: function() {
				this.enabled = true;
			},
			disable: function() {
				this.enabled = false;
			},
			destroy: function() {
				var dom = this.el.clone();
				var p = this.container.parent();
				this.container.remove();
				dom.removeAttr('style');
				p.append(dom);
			}
		};

		var detectGestures = function(el, target, viewerContainer) {
			var interaction = null,
				fingers = 0,
				lastTouchStart = null,
				startTouches = null,
				lastTouchY = null,
				clientY = null,
				lastclientY = 0,
				lastTop = 0,
				setInteraction = function(newInteraction, event) {
					if (interaction !== newInteraction) {

						if (interaction && !newInteraction) {
							switch (interaction) {
								case "zoom":
									target.handleZoomEnd(event);
									break;
								case 'drag':
									target.handleDragEnd(event);
									break;
							}
						}

						switch (newInteraction) {
							case 'zoom':
								target.handleZoomStart(event);
								break;
							case 'drag':
								target.handleDragStart(event);
								break;
						}
					}
					interaction = newInteraction;
				},

				updateInteraction = function(event) {
					if (fingers === 2) {
						setInteraction('zoom');
					} else if (fingers === 1 && target.canDrag()) {
						setInteraction('drag', event);
					} else {
						setInteraction(null, event);
					}
				},

				targetTouches = function(touches) {
					return Array.prototype.slice.call(touches).map(function(touch) {
						return {
							x: touch.pageX,
							y: touch.pageY
						};
					});
				},

				getDistance = function(a, b) {
					var x, y;
					x = a.x - b.x;
					y = a.y - b.y;
					return Math.sqrt(x * x + y * y);
				},

				calculateScale = function(startTouches, endTouches) {
					var startDistance = getDistance(startTouches[0], startTouches[1]),
						endDistance = getDistance(endTouches[0], endTouches[1]);
					return endDistance / startDistance;
				},

				cancelEvent = function(event) {
					event.stopPropagation();
					event.preventDefault();
				},

				detectDoubleTap = function(event) {
					var time = (new Date()).getTime();
					var pageY = event.changedTouches[0].pageY;
					var top = parentNode.scrollTop || 0;
					if (fingers > 1) {
						lastTouchStart = null;
						lastTouchY = null;
						cancelEvent(event);
					}

					if (time - lastTouchStart < 300 && Math.abs(pageY - lastTouchY) < 10 && Math.abs(
							lastTop - top) < 10) {
						cancelEvent(event);
						target.handleDoubleTap(event);
						switch (interaction) {
							case "zoom":
								target.handleZoomEnd(event);
								break;
							case 'drag':
								target.handleDragEnd(event);
								break;
						}
					}

					if (fingers === 1) {
						lastTouchStart = time;
						lastTouchY = pageY;
						lastTop = top;
					}
				},
				firstMove = true;
			if (viewerContainer) {
				var parentNode = viewerContainer[0];
			}
			if (parentNode) {
				parentNode.addEventListener('touchstart', function(event) {
					if (target.enabled) {
						firstMove = true;
						fingers = event.touches.length;
						detectDoubleTap(event);
						clientY = event.changedTouches[0].clientY;
						if (fingers > 1) {
							cancelEvent(event);
						}
					}
				});

				parentNode.addEventListener('touchmove', function(event) {
					if (target.enabled) {
						lastclientY = event.changedTouches[0].clientY;
						if (firstMove) {
							updateInteraction(event);
							startTouches = targetTouches(event.touches);
						} else {
							switch (interaction) {
								case 'zoom':
									target.handleZoom(event, calculateScale(startTouches,
										targetTouches(event.touches)));
									break;
								case 'drag':
									target.handleDrag(event);
									break;
							}
							if (interaction) {
								target.update(lastclientY);
							}
						}
						if (fingers > 1) {
							cancelEvent(event);
						}
						firstMove = false;
					}
				});

				parentNode.addEventListener('touchend', function(event) {
					if (target.enabled) {
						fingers = event.touches.length;
						if (fingers > 1) {
							cancelEvent(event);
						}
						updateInteraction(event);
					}
				});
			}

		};
		return PinchZoom;
	};
	var PinchZoom = definePinchZoom($);
	var Pdfh5 = function(dom, options) {
		this.version = version;
		this.container = $(dom);
		this.options = options;
		this.thePDF = null;
		this.totalNum = null;
		this.pages = null;
		this.initTime = 0;
		this.scale = 1.5;
		this.currentNum = 1;
		this.loadedCount = 0;
		this.endTime = 0;
		this.pinchZoom = null;
		this.timer = null;
		this.docWidth = document.documentElement.clientWidth;
		this.winWidth = $(window).width();
		this.cache = {};
		this.eventType = {};
		this.cacheNum = 1;
		this.resizeEvent = false;
		this.cacheData = null;
		this.pdfjsLibPromise = null;
		this.init(options);
	};
	Pdfh5.prototype = {
		init: function(options) {
			var self = this;
			if (this.container[0].pdfLoaded) {
				this.destroy();
			}
			if (options.cMapUrl) {
				pdfjsLib.cMapUrl = options.cMapUrl;
			} else {
				pdfjsLib.cMapUrl = 'https://unpkg.com/pdfjs-dist@2.0.943/cmaps/';
			}
			pdfjsLib.cMapPacked = true;
			pdfjsLib.rangeChunkSize = 65536;
			this.container[0].pdfLoaded = false;
			this.container.addClass("pdfjs")
			this.initTime = new Date().getTime();
			setTimeout(function() {
				var arr1 = self.eventType["scroll"];
				if (arr1 && arr1 instanceof Array) {
					for (var i = 0; i < arr1.length; i++) {
						arr1[i] && arr1[i].call(self, self.initTime)
					}
				}
			}, 0)
			this.options = this.options ? this.options : {};
			this.options.pdfurl = this.options.pdfurl ? this.options.pdfurl : null;
			this.options.data = this.options.data ? this.options.data : null;
			this.options.scale = this.options.scale ? this.options.scale : this.scale;
			this.options.zoomEnable = this.options.zoomEnable === false ? false : true;
			this.options.scrollEnable = this.options.scrollEnable === false ? false : true;
			this.options.loadingBar = this.options.loadingBar === false ? false : true;
			this.options.pageNum = this.options.pageNum === false ? false : true;
			this.options.backTop = this.options.backTop === false ? false : true;
			this.options.URIenable = this.options.URIenable === true ? true : false;
			this.options.fullscreen = this.options.fullscreen === false ? false : true;
			this.options.lazy = this.options.lazy === true ? true : false;
			this.options.renderType = this.options.renderType === "svg" ? "svg" : "canvas";
			this.options.resize = this.options.resize === false ? false : true;
			this.options.textLayer = this.options.textLayer === true ? true : false;
			this.options.goto = isNaN(this.options.goto) ? 0 : this.options.goto;
			if (this.options.logo && Object.prototype.toString.call(this.options.logo) ===
				'[object Object]' && this.options.logo
				.src) {
				this.options.logo.img = new Image();
				this.options.logo.img.src = this.options.logo.src;
				this.options.logo.img.style.display = "none";
				document.body.appendChild(this.options.logo.img)
			} else {
				this.options.logo = false;
			}
			if (!(this.options.background && (this.options.background.color || this.options.background
					.image))) {
				this.options.background = false
			}
			if (this.options.limit) {
				var n = parseFloat(this.options.limit)
				this.options.limit = isNaN(n) ? 0 : n < 0 ? 0 : n;
			} else {
				this.options.limit = 0
			}
			this.options.type = this.options.type === "fetch" ? "fetch" : "ajax";
			var html = '<div class="loadingBar">' +
				'<div class="progress">' +
				' <div class="glimmer">' +
				'</div>' +
				' </div>' +
				'</div>' +
				'<div class="pageNum">' +
				'<div class="pageNum-bg"></div>' +
				' <div class="pageNum-num">' +
				' <span class="pageNow">1</span>/' +
				'<span class="pageTotal">1</span>' +
				'</div>' +
				' </div>' +
				'<div class="backTop">' +
				'</div>' +
				'<div class="loadEffect loading"></div>';
			if (!this.container.find('.pageNum')[0]) {
				this.container.append(html);
			}
			var viewer = document.createElement("div");
			viewer.className = 'pdfViewer';
			var viewerContainer = document.createElement("div");
			viewerContainer.className = 'viewerContainer';
			viewerContainer.appendChild(viewer);
			this.container.append(viewerContainer);
			this.viewer = $(viewer);
			this.viewerContainer = $(viewerContainer);
			this.pageNum = this.container.find('.pageNum');
			this.pageNow = this.pageNum.find('.pageNow');
			this.pageTotal = this.pageNum.find('.pageTotal');
			this.loadingBar = this.container.find('.loadingBar');
			this.progress = this.loadingBar.find('.progress');
			this.progressDom = this.progress[0];
			this.backTop = this.container.find('.backTop');
			this.loading = this.container.find('.loading');
			if (!this.options.loadingBar) {
				this.loadingBar.hide()
			}
			var containerH = this.container.height(),
				height = containerH * (1 / 3);

			if (!this.options.scrollEnable) {
				this.viewerContainer.css({
					"overflow": "hidden"
				})
			} else {
				this.viewerContainer.css({
					"overflow": "auto"
				})
			}
			viewerContainer.addEventListener('scroll', function() {
				var scrollTop = viewerContainer.scrollTop;
				if (scrollTop >= 150) {
					if (self.options.backTop) {
						self.backTop.show();
					}
				} else {
					if (self.options.backTop) {
						self.backTop.fadeOut(200);
					}
				}
				if (self.viewerContainer) {
					self.pages = self.viewerContainer.find('.pageContainer');
				}
				clearTimeout(self.timer);
				if (self.options.pageNum && self.pageNum) {
					self.pageNum.show();
				}
				var h = containerH;
				if (self.pages) {
					self.pages.each(function(index, obj) {
						var top = obj.getBoundingClientRect().top;
						var bottom = obj.getBoundingClientRect().bottom;
						if (top <= height && bottom > height) {
							if (self.options.pageNum) {
								self.pageNow.text(index + 1)
							}
							self.currentNum = index + 1;
						}
						if (top <= h && bottom > h) {
							self.cacheNum = index + 1;
						}
					})
				}
				if (scrollTop + self.container.height() >= self.viewer[0].offsetHeight) {
					self.pageNow.text(self.totalNum)
				}
				if (scrollTop === 0) {
					self.pageNow.text(1)
				}
				self.timer = setTimeout(function() {
					if (self.options.pageNum && self.pageNum) {
						self.pageNum.fadeOut(200);
					}
				}, 1500)
				if (self.options.lazy) {
					var num = Math.floor(100 / self.totalNum).toFixed(2);
					if (self.cache[self.cacheNum + ""] && !self.cache[self.cacheNum + ""].loaded) {
						var page = self.cache[self.cacheNum + ""].page;
						var container = self.cache[self.cacheNum + ""].container;
						var pageNum = self.cacheNum;
						self.cache[pageNum + ""].loaded = true;
						var scaledViewport = self.cache[pageNum + ""].scaledViewport;
						if (self.options.renderType === "svg") {
							self.renderSvg(page, scaledViewport, pageNum, num, container, self
								.options)
						} else {
							self.renderCanvas(page, scaledViewport, pageNum, num, container, self
								.options)
						}
					}
					if (self.cache[(self.totalNum - 1) + ""] && self.cache[(self.totalNum - 1) + ""]
						.loaded && !self.cache[self.totalNum +
							""].loaded) {
						var page = self.cache[self.totalNum + ""].page;
						var container = self.cache[self.totalNum + ""].container;
						var pageNum = self.totalNum;
						self.cache[pageNum + ""].loaded = true;
						var scaledViewport = self.cache[pageNum + ""].scaledViewport;
						if (self.options.renderType === "svg") {
							self.renderSvg(page, scaledViewport, pageNum, num, container, self
								.options)
						} else {
							self.renderCanvas(page, scaledViewport, pageNum, num, container, self
								.options)
						}
					}
				}
				var arr1 = self.eventType["scroll"];
				if (arr1 && arr1 instanceof Array) {
					for (var i = 0; i < arr1.length; i++) {
						arr1[i] && arr1[i].call(self, scrollTop, self.currentNum)
					}
				}
			})
			this.backTop.on('click tap', function() {
				var mart = self.viewer.css('transform');
				var arr = mart.replace(/[a-z\(\)\s]/g, '').split(',');
				var s1 = arr[0];
				var s2 = arr[3];
				var x = arr[4] / 2;
				var left = self.viewer[0].getBoundingClientRect().left;
				if (left <= -self.docWidth * 2) {
					x = -self.docWidth / 2
				}
				self.viewer.css({
					transform: 'scale(' + s1 + ', ' + s2 + ') translate(' + x + 'px, 0px)'
				})
				if (self.pinchZoom) {
					self.pinchZoom.offset.y = 0;
					self.pinchZoom.lastclientY = 0;
				}
				self.viewerContainer.animate({
					scrollTop: 0
				}, 300)
				var arr1 = self.eventType["backTop"];
				if (arr1 && arr1 instanceof Array) {
					for (var i = 0; i < arr1.length; i++) {
						arr1[i] && arr1[i].call(self)
					}
				}
			})

			function GetQueryString(name) {
				var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)");
				var r = g.location.search.substr(1).match(reg);
				if (r != null) return decodeURIComponent(r[2]);
				return "";
			}
			var pdfurl = GetQueryString("file"),
				url = "";
			if (pdfurl && self.options.URIenable) {
				url = pdfurl
			} else if (self.options.pdfurl) {
				url = self.options.pdfurl
			}
			if (self.options.loadingBar) {
				self.loadingBar.show();
				self.progress.css({
					width: "3%"
				})
			}

			if (url) {
				if (self.options.type === "ajax") {
					$.ajax({
						type: "get",
						mimeType: 'text/plain; charset=x-user-defined',
						url: url,
						success: function(data) {
							var rawLength = data.length;
							var array = [];
							for (i = 0; i < rawLength; i++) {
								array.push(data.charCodeAt(i) & 0xff);
							}
							self.cacheData = array
							self.renderPdf(self.options, {
								data: array
							})
						},
						error: function(err) {
							self.loading.hide()
							var time = new Date().getTime();
							self.endTime = time - self.initTime;
							var arr1 = self.eventType["complete"];
							if (arr1 && arr1 instanceof Array) {
								for (var i = 0; i < arr1.length; i++) {
									arr1[i] && arr1[i].call(self, "error", err.statusText, self
										.endTime)
								}
							}
							var arr2 = self.eventType["error"];
							if (arr2 && arr2 instanceof Array) {
								for (var i = 0; i < arr2.length; i++) {
									arr2[i] && arr2[i].call(self, err.statusText, self.endTime)
								}
							}
							throw Error(err.statusText)
						}
					});
				} else {
					self.renderPdf(self.options, {
						url: url
					})
				}
			} else if (self.options.data) {
				var data = self.options.data;
				if (typeof data === "string" && data != "") {
					var rawLength = data.length;
					var array = [];
					for (i = 0; i < rawLength; i++) {
						array.push(data.charCodeAt(i) & 0xff);
					}
					self.cacheData = array
					self.renderPdf(self.options, {
						data: array
					})
				} else if (typeof data === "object") {
					if (data.length == 0) {
						var time = new Date().getTime();
						self.endTime = time - self.initTime;
						var arr1 = self.eventType["complete"];
						if (arr1 && arr1 instanceof Array) {
							for (var i = 0; i < arr1.length; i++) {
								arr1[i] && arr1[i].call(self, "error", "options.data is empty Array", self
									.endTime)
							}
						}
						var arr2 = self.eventType["error"];
						if (arr2 && arr2 instanceof Array) {
							for (var i = 0; i < arr2.length; i++) {
								arr2[i] && arr2[i].call(self, "options.data is empty Array", self.endTime)
							}
						}
						throw Error("options.data is empty Array")
					} else {
						self.cacheData = data
						self.renderPdf(self.options, {
							data: data
						})
					}
				}

			} else {
				var time = new Date().getTime();
				self.endTime = time - self.initTime;
				var arr1 = self.eventType["complete"];
				if (arr1 && arr1 instanceof Array) {
					for (var i = 0; i < arr1.length; i++) {
						arr1[i] && arr1[i].call(self, "error", "Expect options.pdfurl or options.data!",
							self.endTime)
					}
				}
				var arr2 = self.eventType["error"];
				if (arr2 && arr2 instanceof Array) {
					for (var i = 0; i < arr2.length; i++) {
						arr2[i] && arr2[i].call(self, "Expect options.pdfurl or options.data!", self
							.endTime)
					}
				}
				throw Error("Expect options.pdfurl or options.data!")
			}
		},
		renderPdf: function(options, obj) {
			this.container[0].pdfLoaded = true;
			var self = this;
			if (options.cMapUrl) {
				obj.cMapUrl = options.cMapUrl;
			} else {
				// obj.cMapUrl = 'https://unpkg.com/pdfjs-dist@2.0.943/cmaps/';
				obj.cMapUrl = "https://www.gjtool.cn/cmaps/"
			}
			if (options.httpHeaders) {
				obj.httpHeaders = options.httpHeaders;
			}
			if (options.withCredentials) {
				obj.withCredentials = true;
			}
			if (options.password) {
				obj.password = options.password;
				console.log(obj.password)
			}
			if (options.stopAtErrors) {
				obj.stopAtErrors = true;
			}
			if (options.disableFontFace) {
				obj.disableFontFace = true;
			}
			if (options.disableRange) {
				obj.disableRange = true;
			}
			if (options.disableStream) {
				obj.disableStream = true;
			}
			if (options.disableAutoFetch) {
				obj.disableAutoFetch = true;
			}
			obj.cMapPacked = true;
			obj.rangeChunkSize = 65536;
			this.pdfjsLibPromise = pdfjsLib.getDocument(obj).then(function(pdf) {
				self.loading.hide()
				self.thePDF = pdf;
				self.totalNum = pdf.numPages;
				if (options.limit > 0) {
					self.totalNum = options.limit
				}
				self.pageTotal.text(self.totalNum)
				if (!self.pinchZoom) {
					var arr1 = self.eventType["ready"];
					if (arr1 && arr1 instanceof Array) {
						for (var i = 0; i < arr1.length; i++) {
							arr1[i] && arr1[i].call(self)
						}
					}
					self.pinchZoom = new PinchZoom(self.viewer, {
						tapZoomFactor: options.tapZoomFactor,
						zoomOutFactor: options.zoomOutFactor,
						animationDuration: options.animationDuration,
						maxZoom: options.maxZoom,
						minZoom: options.minZoom
					}, self.viewerContainer);
					var timeout, firstZoom = true;
					self.pinchZoom.done = function(scale) {
						clearTimeout(timeout)
						timeout = setTimeout(function() {
							if (self.options.renderType === "svg") {
								return
							}
							if (scale <= 1 || self.options.scale == 5) {
								return
							}
							if (self.thePDF) {
								self.thePDF.destroy();
								self.thePDF = null;
							}
							self.options.scale = scale;
							self.renderPdf(self.options, {
								data: self.cacheData
							})
						}, 310)
						if (scale == 1) {
							if (self.viewerContainer) {
								self.viewerContainer.css({
									'-webkit-overflow-scrolling': 'touch'
								})
							}

						} else {
							if (self.viewerContainer) {
								self.viewerContainer.css({
									'-webkit-overflow-scrolling': 'auto'
								})
							}
						}
						var arr1 = self.eventType["zoom"];
						if (arr1 && arr1 instanceof Array) {
							for (var i = 0; i < arr1.length; i++) {
								arr1[i] && arr1[i].call(self, scale)
							}
						}
					}
					if (options.zoomEnable) {
						self.pinchZoom.enable()
					} else {
						self.pinchZoom.disable()
					}
				}

				var promise = Promise.resolve();
				var num = Math.floor(100 / self.totalNum).toFixed(2);
				var i = 1;
				for (i = 1; i <= self.totalNum; i++) {
					self.cache[i + ""] = {
						page: null,
						loaded: false,
						container: null,
						scaledViewport: null
					};
					promise = promise.then(function(pageNum) {
						return self.thePDF.getPage(pageNum).then(function(page) {
							setTimeout(function() {
								if (self.options.goto) {
									if (pageNum == self.options.goto) {
										self.goto(pageNum)
									}
								}
							}, 0)

							self.cache[pageNum + ""].page = page;
							var viewport = page.getViewport(options.scale);
							var scale = (self.docWidth / viewport.width).toFixed(2)
							var scaledViewport = page.getViewport(parseFloat(scale))
							var div = self.container.find('.pageContainer' +
								pageNum)[0];
							var container;
							if (!div) {
								container = document.createElement('div');
								container.className =
									'pageContainer pageContainer' + pageNum;
								container.setAttribute('name', 'page=' + pageNum);
								container.setAttribute('title', 'Page ' + pageNum);
								var loadEffect = document.createElement('div');
								loadEffect.className = 'loadEffect';
								container.appendChild(loadEffect);
								self.viewer[0].appendChild(container);
								if (window.ActiveXObject || "ActiveXObject" in
									window) {
									$(container).css({
										'width': viewport.width + 'px',
										"height": viewport.height + 'px'
									}).attr("data-scale", viewport.width /
										viewport.height)
								} else {
									var h = $(container).width() / (viewport
										.viewBox[2] / viewport.viewBox[3]);
									if (h > viewport.height) {
										h = viewport.height
									}
									$(container).css({
										'max-width': viewport.width,
										"max-height": viewport.height,
										"min-height": h + 'px'
									}).attr("data-scale", viewport.width /
										viewport.height)
								}
							} else {
								container = div
							}
							if (options.background) {
								/*背景颜色*/
								if (options.background.color) {
									container.style["background-color"] = options
										.background.color
								}
								/*背景图片*/
								if (options.background.image) {
									container.style["background-image"] = options
										.background.image
								}
								/*平铺与否*/
								if (options.background.repeat) {
									container.style["background-repeat"] = options
										.background.repeat
								}
								/*背景图片位置*/
								if (options.background.position) {
									container.style["background-position"] = options
										.background.position
								}
								/*背景图像的尺寸*/
								if (options.background.size) {
									container.style["background-size"] = options
										.background.size
								}
							}
							self.cache[pageNum + ""].container = container;
							self.cache[pageNum + ""].scaledViewport =
								scaledViewport;
							var sum = 0,
								containerH = self.container.height();
							self.pages = self.viewerContainer.find(
								'.pageContainer');
							if (options.resize) {
								self.resize()
							}
							if (self.pages && options.lazy) {
								self.pages.each(function(index, obj) {
									var top = obj.offsetTop;
									if (top <= containerH) {
										sum = index + 1;
										self.cache[sum + ""].loaded = true;
									}
								})
							}

							if (pageNum > sum && options.lazy) {
								return
							}
							if (options.renderType === "svg") {
								return self.renderSvg(page, scaledViewport, pageNum,
									num, container, options, viewport)
							}
							return self.renderCanvas(page, scaledViewport, pageNum,
								num, container, options)
						});
					}.bind(null, i));
				}
			}).catch(function(err) {
				self.loading.hide();
				var time = new Date().getTime();
				self.endTime = time - self.initTime;
				var arr1 = self.eventType["complete"];
				if (arr1 && arr1 instanceof Array) {
					for (var i = 0; i < arr1.length; i++) {
						arr1[i] && arr1[i].call(self, "error", err.message, self.endTime)
					}
				}
				var arr2 = self.eventType["error"];
				if (arr2 && arr2 instanceof Array) {
					for (var i = 0; i < arr2.length; i++) {
						arr2[i] && arr2[i].call(self, err.message, self.endTime)
					}
				}
			})
		},
		renderSvg: function(page, scaledViewport, pageNum, num, container, options, viewport) {
			var self = this;
			var viewport = page.getViewport(options.scale);
			var scale = (self.docWidth / viewport.width).toFixed(2)
			return page.getOperatorList().then(function(opList) {
				var svgGfx = new pdfjsLib.SVGGraphics(page.commonObjs, page.objs);
				return svgGfx.getSVG(opList, scaledViewport).then(function(svg) {
					self.loadedCount++;
					container.children[0].style.display = "none";
					container.appendChild(svg);
					svg.style.width = "100%";
					svg.style.height = "100%";
					if (self.options.loadingBar) {
						self.progress.css({
							width: num * self.loadedCount + "%"
						})
					}
					var time = new Date().getTime();
					var arr1 = self.eventType["render"];
					if (arr1 && arr1 instanceof Array) {
						for (var i = 0; i < arr1.length; i++) {
							arr1[i] && arr1[i].call(self, pageNum, time - self.initTime,
								container)
						}
					}
					if (self.loadedCount === self.totalNum) {
						self.finalRender(options)
					}
				});
			}).then(function() {
				return page.getTextContent();
			}).then(function(textContent) {
				if (!self.options.textLayer) {
					return
				}
				if ($(container).find(".textLayer")[0]) {
					return
				}
				var textLayerDiv = document.createElement('div');
				textLayerDiv.setAttribute('class', 'textLayer');
				container.appendChild(textLayerDiv);
				viewport.width = viewport.width * scale;
				viewport.height = viewport.height * scale;
				var textLayer = new TextLayerBuilder({
					textLayerDiv: textLayerDiv,
					pageIndex: page.pageIndex,
					viewport: viewport
				});

				textLayer.setTextContent(textContent);

				textLayer.render();
			});;
		},
		renderCanvas: function(page, viewport, pageNum, num, container, options) {
			var self = this;
			var viewport = page.getViewport(options.scale);
			var scale = (self.docWidth / viewport.width).toFixed(2)
			var canvas = document.createElement("canvas");
			var obj2 = {
				'Cheight': viewport.height * scale,
				'width': viewport.width,
				'height': viewport.height,
				'canvas': canvas,
				'index': self.loadedCount
			}
			var context = canvas.getContext('2d');
			if (options.logo) {
				context.drawImage(self.options.logo.img, self.options.logo.x * self.options.scale,
					self.options.logo.y * self.options.scale, self.options.logo.width * self.options
					.scale, self.options.logo.height *
					self.options.scale
				);
			}
			canvas.height = viewport.height;
			canvas.width = viewport.width;
			if (self.options.loadingBar) {
				self.progress.css({
					width: num * self.loadedCount + "%"
				})
			}
			obj2.src = obj2.canvas.toDataURL("image/png");
			var renderObj = {
				canvasContext: context,
				viewport: viewport
			}
			if (options.background) {
				renderObj.background = "rgba(255, 255, 255, 0)"
			}
			return page.render(renderObj).then(function() {
				if (options.logo) {
					context.drawImage(self.options.logo.img, self.options.logo.x * self.options
						.scale,
						self.options.logo.y * self.options.scale, self.options.logo.width * self
						.options.scale, self.options.logo.height *
						self.options.scale
					);
				}
				self.loadedCount++;
				var img = new Image();
				var time = new Date().getTime();
				var time2 = 0;
				if (self.renderTime == 0) {
					time2 = time - self.startTime
				} else {
					time2 = time - self.renderTime
				}
				obj2.src = obj2.canvas.toDataURL("image/png");

				img.src = obj2.src;
				img.className = "canvasImg" + pageNum;
				var img0 = self.container.find(".pageContainer" + pageNum).find(".canvasImg" +
					pageNum)[0];
				if (container && !img0) {
					container.appendChild(img);
				} else if (img0) {
					img0.src = obj2.src
				}
				container.children[0].style.display = "none";
				var time = new Date().getTime();
				var arr1 = self.eventType["render"];
				if (arr1 && arr1 instanceof Array) {
					for (var i = 0; i < arr1.length; i++) {
						arr1[i] && arr1[i].call(self, pageNum, time - self.initTime, container)
					}
				}
				if (self.loadedCount === self.totalNum) {
					self.finalRender(options)
				}
			}).then(function() {
				return page.getTextContent();
			}).then(function(textContent) {
				if (!self.options.textLayer) {
					return
				}
				if ($(container).find(".textLayer")[0]) {
					return
				}
				var textLayerDiv = document.createElement('div');
				textLayerDiv.setAttribute('class', 'textLayer');
				container.appendChild(textLayerDiv);
				viewport.width = viewport.width * scale;
				viewport.height = viewport.height * scale;
				var textLayer = new TextLayerBuilder({
					textLayerDiv: textLayerDiv,
					pageIndex: page.pageIndex,
					viewport: viewport
				});

				textLayer.setTextContent(textContent);

				textLayer.render();
			});
		},
		finalRender: function(options) {
			var time = new Date().getTime();
			var self = this;
			if (self.options.loadingBar) {
				self.progress.css({
					width: "100%"
				});
			}
			setTimeout(function() {
				self.loadingBar.hide();
			}, 300)
			self.endTime = time - self.initTime;
			if (options.renderType === "svg") {
				if (self.totalNum !== 1) {
					self.cache[(self.totalNum - 1) + ""].loaded = true;
				} else {
					self.cache["1"].loaded = true;
				}
			}
			if (options.zoomEnable) {
				if (self.pinchZoom) {
					self.pinchZoom.enable()
				}
			} else {
				if (self.pinchZoom) {
					self.pinchZoom.disable()
				}
			}
			var arr1 = self.eventType["complete"];
			if (arr1 && arr1 instanceof Array) {
				for (var i = 0; i < arr1.length; i++) {
					arr1[i] && arr1[i].call(self, "success", "pdf加载完成", self.endTime)
				}
			}
			var arr2 = self.eventType["success"];
			if (arr2 && arr2 instanceof Array) {
				for (var i = 0; i < arr2.length; i++) {
					arr2[i] && arr2[i].call(self, self.endTime)
				}
			}
		},
		resize: function() {
			var self = this;
			if (self.resizeEvent) {
				return
			}
			self.resizeEvent = true;
			var timer;
			if (self.pages) {
				$(window).on("resize", function() {
					self.pages.each(function(i, item) {
						$(item).css("min-height", "auto")
					})
				})
			}
		},
		show: function(callback) {
			this.container.show();
			callback && callback.call(this)
			var arr = this.eventType["show"];
			if (arr && arr instanceof Array) {
				for (var i = 0; i < arr.length; i++) {
					arr[i] && arr[i].call(this)
				}
			}
		},
		hide: function(callback) {
			this.container.hide()
			callback && callback.call(this)
			var arr = this.eventType["hide"];
			if (arr && arr instanceof Array) {
				for (var i = 0; i < arr.length; i++) {
					arr[i] && arr[i].call(this)
				}
			}
		},
		on: function(type, callback) {
			if (this.eventType[type] && this.eventType[type] instanceof Array) {
				this.eventType[type].push(callback)
			}
			this.eventType[type] = [callback]
		},
		off: function(type) {
			if (type !== undefined) {
				this.eventType[type] = [null]
			} else {
				for (var i in this.eventType) {
					this.eventType[i] = [null]
				}
			}
		},
		goto: function(num) {
			var self = this;

			if (!isNaN(num)) {
				if (self.viewerContainer) {
					self.pages = self.viewerContainer.find('.pageContainer');

					if (self.pages) {
						var h = 0;
						var signHeight = 0;
						if (num - 1 > 0) {
							signHeight = self.pages[0].getBoundingClientRect().height;
						}
						self.viewerContainer.animate({
							scrollTop: signHeight * (num - 1) + 8 * num
						}, 300)
					}
				}
			}
		},
		scrollEnable: function(flag) {
			if (flag === false) {
				this.viewerContainer.css({
					"overflow": "hidden"
				})
			} else {
				this.viewerContainer.css({
					"overflow": "auto"
				})
			}
			var arr = this.eventType["scrollEnable"];
			if (arr && arr instanceof Array) {
				for (var i = 0; i < arr.length; i++) {
					arr[i] && arr[i].call(this, flag)
				}
			}
		},
		zoomEnable: function(flag) {
			if (!this.pinchZoom) {
				return
			}
			if (flag === false) {
				this.pinchZoom.disable()
			} else {
				this.pinchZoom.enable()
			}
			var arr = this.eventType["zoomEnable"];
			if (arr && arr instanceof Array) {
				for (var i = 0; i < arr.length; i++) {
					arr[i] && arr[i].call(this, flag)
				}
			}
		},
		download: function(name, callback) {
			if (this.options.pdfurl) {
				download(this.options.pdfurl, name, callback)
			} else if (this.options.data) {
				fileDownLoad(this.options.data, name, callback)
			}
		},
		reset: function(callback) {
			if (this.pinchZoom) {
				this.pinchZoom.offset.y = 0;
				this.pinchZoom.offset.x = 0;
				this.pinchZoom.lastclientY = 0;
				this.pinchZoom.zoomFactor = 1;
				this.pinchZoom.update();
			}
			if (this.viewerContainer) {
				this.viewerContainer.scrollTop(0);
			}
			callback && callback.call(this)
			var arr = this.eventType["reset"];
			if (arr && arr instanceof Array) {
				for (var i = 0; i < arr.length; i++) {
					arr[i] && arr[i].call(this)
				}
			}
		},
		destroy: function(callback) {
			this.reset();
			this.off();
			if (this.thePDF) {
				this.thePDF.destroy();
				this.thePDF = null;
			}
			if (this.viewerContainer) {
				this.viewerContainer.remove();
				this.viewerContainer = null;
			}
			if (this.container) {
				this.container.html('');
			}
			this.totalNum = null;
			this.pages = null;
			this.initTime = 0;
			this.endTime = 0;
			this.viewer = null;
			this.pageNum = null;
			this.pageNow = null;
			this.pageTotal = null;
			this.loadingBar = null;
			this.progress = null;
			this.loadedCount = 0;
			this.timer = null;
			callback && callback.call(this)
			var arr = this.eventType["destroy"];
			if (arr && arr instanceof Array) {
				for (var i = 0; i < arr.length; i++) {
					arr[i] && arr[i].call(this)
				}
			}
		}
	}
	return Pdfh5;

	function download(url, name, callback) {
		var xhr = new XMLHttpRequest();
		xhr.open('GET', url, true);
		xhr.responseType = "blob";
		if (Object.prototype.toString.call(name) === "[object Function]") {
			callback = name
			name = undefined;
		}
		name = name ? name : "download.pdf";
		if (name.indexOf(".pdf") == -1) {
			name += ".pdf"
		}
		xhr.onload = function() {
			if (this.status === 200) {
				var blob = this.response;
				var reader = new FileReader();
				reader.readAsDataURL(blob);
				reader.onload = function(e) {
					var a = document.createElement('a');
					a.download = name;
					a.href = e.target.result;
					$("body").append(a);
					a.click();
					$(a).remove();
					callback && callback()
				}
			}
		};
		xhr.send()
	}

	function fileDownLoad(data, name, callback) {
		if (Object.prototype.toString.call(name) === "[object Function]") {
			callback = name
			name = undefined;
		}
		name = name ? name : "download.pdf";
		if (name.indexOf(".pdf") == -1) {
			name += ".pdf"
		}
		var array = null
		try {
			var enc = new TextDecoder('utf-8')
			array = JSON.parse(enc.decode(new Uint8Array(data)))
		} catch (err) {
			if (Object.prototype.toString.call(data) === "[object ArrayBuffer]") {
				array = data
			} else {
				if (Object.prototype.toString.call(data) === "[object Array]") {
					array = new Uint8Array(data);
				} else {
					var rawLength = data.length;
					array = new Uint8Array(new ArrayBuffer(rawLength));
				}
				for (var i = 0; i < rawLength; i++) {
					array[i] = data.charCodeAt(i) & 0xff;
				}
			}
			var blob = new Blob([array]);
			var a = document.createElement('a');
			var url = window.URL.createObjectURL(blob);
			a.download = name;
			a.href = url;
			$("body").append(a);
			a.click();
			$(a).remove();
			callback && callback()
		}

	}
});
