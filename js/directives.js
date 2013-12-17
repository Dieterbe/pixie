'use strict';


angular.module('directives', []).
directive('keybinding', function () {
    return {
        restrict: 'E',
        scope: {
            invoke: '&'
        },
        link: function (scope, el, attr) {
            Mousetrap.bind(attr.on, scope.invoke);
        }
    };
}).
 // http://stackoverflow.com/questions/12790854/angular-directive-to-scroll-to-a-given-item
directive('scrollIf', function () {
    console.log("scroll top level");
    return {
        link: function (scope, element, attributes) {
            console.log("scroll function cb");
            setTimeout(function () {
                scope.$watch(attributes.scrollIf, function(val) {
                    if (val) {
                        console.debug("scrolling to #photo-" + scope.$parent.focusIndex);
                        //window.scrollTo(0, element.offsetTop - 200)
                        //window.scrollTo(0, $("#photo-"+ scope.$parent.focusIndex).offsetTop);
                        window.scrollTo(0, document.getElementById("photo-"+ scope.$parent.focusIndex).offsetTop - 200);
                    }
                });
            });
        }
    }
});
 

