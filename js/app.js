'use strict';

/* App Module */

var photosApp = angular.module('photosApp', [
  'ngRoute',
  'photosControllers',
  'photosFilters',
  'photosServices'
]);

photosApp.config(['$routeProvider',
  function($routeProvider) {
    $routeProvider.
      when('/photos', {
        templateUrl: 'partials/photos.html',
        controller: 'PhotosCtrl'
      }).
      otherwise({
        redirectTo: '/photos'
      });
  }]);
