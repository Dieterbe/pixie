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
        console.log('opening : ', $scope.photos[$scope.focusIndex] );
    };
    $scope.moveDown = function () {
        if ($scope.focusIndex < $scope.photos.length -1) {
            $scope.focusIndex++;
            console.log($scope.focusIndex);
        }
    }
    $scope.moveUp = function () {
        if($scope.focusIndex > 0 ) {
            $scope.focusIndex--;
            console.log($scope.focusIndex);
        }
    }
  }]);
