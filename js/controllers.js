'use strict';

/* Controllers */

var photosControllers = angular.module('photosControllers', []);

photosControllers.controller('PhotosCtrl', ['$scope', '$routeParams', 'Photos',
  function($scope, $routeParams, Photos) {
    $scope.setDirectory = function() {
      $scope.photos = Photos.get({directory: $scope.directory});
    }
  }]);
