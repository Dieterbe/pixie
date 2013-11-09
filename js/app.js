'use strict';

/* App Module */

var photosApp = angular.module('photosApp', [
  'ngRoute',
  'photosControllers',
  'photosFilters',
  'photosServices',
  'directives'
]);

photosApp.config(['$routeProvider',
  function($routeProvider) {
    $routeProvider.
      when('/:dir', {
        templateUrl: 'partials/photos.html',
        controller: 'PhotosCtrl'
      }).
      otherwise({
        redirectTo: '/'
      });
  }]);
