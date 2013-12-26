'use strict';


angular.module('directives', []).
 // http://stackoverflow.com/questions/12790854/angular-directive-to-scroll-to-a-given-item
directive('scrollIf', function () {
    return {
        link: function (scope, element, attributes) {
            setTimeout(function () {
                scope.$watch(attributes.scrollIf, function(val) {
                    if (val) {
                        //window.scrollTo(0, element.offsetTop - 200)
                        //window.scrollTo(0, $("#photo-"+ scope.$parent.focusIndex).offsetTop);
                        window.scrollTo(0, document.getElementById("photo-"+ scope.$parent.focusIndex).offsetTop - 200);
                    }
                });
            });
        }
    }
});
 

