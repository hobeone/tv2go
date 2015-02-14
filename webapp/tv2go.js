angular.module('tv2go', [
  'ngAnimate',
  'ui.router',
  'shows',
  'shows.episodes',
])
.config(function($stateProvider, $urlRouterProvider){
  $urlRouterProvider.otherwise("/");

  $stateProvider
  .state('tv2go', {
    url: '',
    abstract: true
  });
})
;
