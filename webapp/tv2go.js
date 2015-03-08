angular.module('tv2go', [
  'ngAnimate',
  'ui.router',
  'ng.group',
  'xeditable',
  'ui.bootstrap',
  'angularModalService',
  'shows',
  'shows.episodes',
])
.config(function($stateProvider, $urlRouterProvider){
  'use strict';
  $urlRouterProvider.otherwise('/shows');

  $stateProvider
  .state('tv2go', {
    templateUrl: 'layout.tmpl.html',
    controller: 'IndexCtrl as indexCtrl',
    abstract: true,
  });
})
.run(function(editableOptions) {
  editableOptions.theme = 'bs3';
})
.filter('humanize', function(){
    return function humanize(number) {
        if(number < 1000) {
            return number;
        }
        var si = ['K', 'M', 'G', 'T', 'P', 'H'];
        var exp = Math.floor(Math.log(number) / Math.log(1000));
        var result = number / Math.pow(1000, exp);
        result = (result % 1 > (1 / Math.pow(1000, exp - 1))) ? result.toFixed(2) : result.toFixed(0);
        return result + si[exp - 1];
    };
})
.filter('humanizeTime', function() {
  return function(t) {
    // GO json time format:
    return moment(t, "YYYY-MM-DDTHH:mm:ssZ").fromNow();
  }
})
.filter('nextAirDate', function() {
  return function(t) {
    if(angular.isDefined(t)) {
      // GO json time format:
      return moment(t, "YYYY-MM-DDTHH:mm:ssZ").format("L");
    }
    return "";
  }
})
.filter('nextAirDateHumanize', function() {
  return function(t) {
    if(angular.isDefined(t)) {
      // GO json time format:
      return moment(t, "YYYY-MM-DDTHH:mm:ssZ").calendar();
    }
    return "";
  }
})
.controller('IndexCtrl', function(){});
