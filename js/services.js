'use strict';

var photosServices = angular.module('photosServices', ['ngResource']);

photosServices.factory('Photos', ['$resource',
  function($resource){
    return $resource('/api/photos/dir=:directory', {}, {
	    get: { method:'GET', params:{directory:'photos'}, isArray:true}}
    );
  }]);
