'use strict';

/* Controllers */

var photosControllers = angular.module('photosControllers', []);

photosControllers.controller('PhotosCtrl', ['$scope', '$routeParams', 'Photos',
  function($scope, $routeParams, Photos) {
    $scope.setDirectory = function() {
        $scope.photos = Photos.get({directory: $scope.directory});
    }
    $scope.focusIndex = 0;
    $scope.openRecord = function () {
	    $scope.$apply(function () {
        console.log('opening : ', $scope.photos[$scope.focusIndex] );
	    });
    };
    $scope.moveDown = function () {
	    $scope.$apply(function () {
            if ($scope.focusIndex < $scope.photos.length -1) {
                $scope.focusIndex++;
                window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
            }
	    });
    }
    $scope.moveUp = function () {
	    $scope.$apply(function () {
            if($scope.focusIndex > 0 ) {
                $scope.focusIndex--;
                window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
            }
	    });
    }
  }]);
