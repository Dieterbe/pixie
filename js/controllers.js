'use strict';

/* Controllers */

var photosControllers = angular.module('photosControllers', []);

photosControllers.controller('PhotosCtrl', ['$scope', '$routeParams', 'Binds', 'Photos', 'Photo', 'Edit', '$timeout',
  function($scope, $routeParams, Binds, Photos, Photo, Edit, $timeout) {
    Binds.list(
        function(response) {
            // TODO: json validation!
            //console.debug(response);
            $.map(response, function (v, k) {
                // ignore object built-ins, we only want the actual json key-values
                if (typeof(v) == "string") {
                    // later we may want to do this properly will apply() or smth
                    Mousetrap.bind(k, function (){ eval("$scope." + v); });
                }
            });
            $scope.logs.push({msg: "keybinds loaded", type: 'info'});
        }, function(response) {
            console.debug(response);
            $scope.logs.push({msg: "could not load keybinds", type: 'error'});
        });

    // routeParams.dir
    $scope.focusIndex = 0; // determines position top-bottom
    $scope.subFocusIndex = 0; // determines position left (original) to right (any edits)
    $scope.viewDistance = 7; // how many pictures to put in the DOM before and after the focused photo
    $scope.logs = [];

    $scope.setDirectory = function() {
        $scope.focusIndex = 0;
        $scope.subFocusIndex = 0;
        $scope.photosAll = Photos.get({directory: $scope.directory}, function(response) {
            $scope.logs.push({msg: "loaded " + $scope.directory, type: 'info'});
            $scope.setPhotosView();
        }, function(response) {
            $scope.logs.push({msg: "failed to load " + $scope.directory, type: 'error'});
            $scope.photosAll = [];
            $scope.setPhotosView();
        });
    }
    $scope.setPhotosView = function () {
        var lower = $scope.focusIndex - $scope.viewDistance;
        if (lower < 0) {
            lower = 0;
        }
        var upper = $scope.focusIndex + $scope.viewDistance + 1;
        if (upper > $scope.photosAll.length) {
            upper = $scope.photosAll.length;
        }
        $scope.photosView = $scope.photosAll.slice(lower, upper);
        console.log("real length " + $scope.photosAll.length + ", FocusIndex " + $scope.focusIndex);
        console.log("win subset lower " + lower + ", upper " + upper);
        console.log("win  length " + $scope.photosView.length);
        var str = "";
        for (var i = 0; i < $scope.photosView.length; i++) {
            str += " " + $scope.photosView[i].id;
        }
        console.log(str);
    }
    $scope.getCurrentPhoto = function() {
        var img = $scope.photosAll[$scope.focusIndex];
        if ($scope.subFocusIndex > 0 ) {
            var idx = 0;
            for (var key in img.edits) {
                idx++;
                console.debug(key, idx);
                if (idx == $scope.subFocusIndex) {
                    img = img.edits[key];
                    break;
                }
            }
        }
        console.debug("returning current photo, ", img);
        return img;
    };
    $scope.openRecord = function () {
        $scope.$apply(function () {
        console.log('opening : ', $scope.photosAll[$scope.focusIndex] );
        });
    };
    $scope.newEdit = function () {
        $scope.$apply(function () {
            console.log('creating edit for :', $scope.photosAll[$scope.focusIndex] );
            $scope.photosAll[$scope.focusIndex] = Edit.new($scope.photosAll[$scope.focusIndex], function(response) {
                $scope.logs.push({msg: "created new edit for " + $scope.photosAll[$scope.focusIndex].name, type: 'info'});
            }, function(response) {
                $scope.logs.push({msg: "failed to edit " + $scope.photosAll[$scope.focusIndex].name, type: 'error'});
            });
            $scope.setPhotosView();
        });
    };
    $scope.moveHome = function () {
        $scope.$apply(function () {
            $scope.focusIndex = 0;
            $scope.subFocusIndex = 0;
            $scope.setPhotosView();
        });
    }
    $scope.moveEnd = function () {
        $scope.$apply(function () {
            $scope.focusIndex = $scope.photosAll.length -1
            $scope.subFocusIndex = 0;
            $scope.setPhotosView();
        });
    }
    $scope.moveDown = function () {
        $scope.$apply(function () {
            if ($scope.focusIndex < $scope.photosAll.length -1) {
                $scope.focusIndex++;
                $scope.subFocusIndex = 0;
                $scope.setPhotosView();
            }
        });
    }
    $scope.moveUp = function () {
        $scope.$apply(function () {
            if($scope.focusIndex > 0 ) {
                $scope.focusIndex--;
                $scope.subFocusIndex = 0;
                $scope.setPhotosView();
            }
        });
    }
    $scope.moveLeft = function () {
        $scope.$apply(function () {
            if($scope.subFocusIndex > 0 ) {
                $scope.subFocusIndex--;
            }
        });
    }
    $scope.moveRight = function () {
        $scope.$apply(function () {
            // 3 edits means subFocusIndex can be max 3 (0 original, 1/2/3 for the edits)
            if($scope.subFocusIndex < Object.keys($scope.photosAll[$scope.focusIndex].edits).length) {
                $scope.subFocusIndex++;
            }
        });
    }
    $scope.tag = function (tag) {
        $scope.$apply(function () {
            // not sure if needed, but by the time the callback fires we may have focused on other image
            var img = $scope.getCurrentPhoto();
            img.tag = tag;
            Photo.tag(img, function(response) {
                $scope.logs.push({msg: response.msg + ": " + img.dir + "/" + img.name + " (" + tag + ")", type: 'info'});
                if(response.msg != "tag already existed") {
                    img['tags'].push(tag);
                }
            }, function(response) {
                $scope.logs.push({msg: response.msg + ": "  + img.dir + "/" + img.name + " (" + tag + ")", type: 'error'});
            });
            $scope.setPhotosView();
        });
    }

    $scope.unTag = function (tag) {
        $scope.$apply(function () {
            // not sure if needed, but by the time the callback fires we may have focused on other image
            var img = $scope.getCurrentPhoto();
            img.tag = tag;
            Photo.untag(img, function(response) {
                console.debug(response);
                $scope.logs.push({msg: response.msg + ": " + img.dir + "/" + img.name + " (" + tag + ")", type: 'info'});
                var tag_index = img['tags'].indexOf(tag)
                if(tag_index!=-1){
                       img['tags'].splice(tag_index, 1);
                }
            }, function(response) {
                console.debug(response);
                $scope.logs.push({msg: response.msg + ": " + img.dir + "/" + img.name + " (" + tag + ")", type: 'error'});
            });
            $scope.setPhotosView();
        });
    }
    // todo $scope.autotag = function (tag) {
  }]);
photosControllers.controller('PhotoDetailCtrl', ['$scope', '$routeParams', 'Photos', 'Photo',
  function($scope, $routeParams, Photos, Photo) {
  }]);
