angular.module('tv2go', [
  'ngAnimate',
  'ui.router',
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
.controller('IndexCtrl', function(){});
