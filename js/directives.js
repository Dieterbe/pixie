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
directive('scrollIfbad', function () {
    console.log("scroll top level");
    return {
        link: function (scope, element, attributes) {
            console.log("scroll function cb");
            setTimeout(function () {
                console.log("scroll in setTimeout running..");
                console.debug(element);
                console.debug(attributes.scrollIf);
                console.debug(scope.photo.id + "==" + scope.focusIndex + "?");
                if (scope.$eval(attributes.scrollIf)) {
                    console.debug("yep");
                    window.scrollTo(0, element.offsetTop - 200)
                }
            });
        }
    }
}).
directive('scrollIfalsobad', function () {
    console.log("scroll top level");
    return function (scope, element, attributes) {
        console.log("scroll function cb");
        setTimeout(function () {
            console.log("scroll in setTimeout running..");
            console.debug(element);
            console.debug(attributes.scrollIf);
            console.debug(scope.photo.id + "==" + scope.focusIndex + "?");
            if (scope.$eval(attributes.scrollIf)) {
                console.debug("scrolling to #photo-" + scope.$parent.focusIndex);
                console.debug($("#photo-" + scope.$parent.focusIndex)[0]);
                //window.scrollTo(0, element[0].offsetTop)
                window.scrollTo(0, $("#photo-" + scope.$parent.focusIndex)[0].offsetTop)
            }
        });
    }
}).
directive('scrollUpdater', function () {
  return function (scope, element, attrs) {
console.debug("scroll updater scope");
console.debug(scope);
console.debug("scroll updater watching:");
console.debug(scope.focusIndex);
      console.debug("scrollUpdater running");
      scope.$watch(scope.focusIndex, function(v) {
          console.debug("scrollUpdater: scrolling to:" + scope.focusIndex);
          window.scrollTo(0, ("#photo-" + scope.focusIndex)[0].offsetTop - 100);
      });
  };
});
 

